package filesystem

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/another-d-mention/unicomplex/datastruct/syncmap"
	"github.com/fsnotify/fsnotify"
)

type base struct {
	userRoot    string // whatever the user-provided root directory is
	rootDir     string // the absolute path to the root directory
	closeNotify chan struct{}
	watcher     *fsnotify.Watcher
	watchers    *syncmap.Map[string, []chan Event]
}

func (b *base) RootDir() string {
	return b.userRoot
}

func (b *base) RootAbs() string {
	return b.rootDir
}

func (b *base) initWatch() error {
	if b.watcher != nil {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	b.closeNotify = make(chan struct{})

	go b.notify()

	b.watcher = watcher
	if b.watchers == nil {
		b.watchers = syncmap.New[string, []chan Event]()
	}

	return nil
}

func (b *base) Watch(path string, callback chan Event) error {
	if err := b.initWatch(); err != nil {
		return err
	}

	path = AbsolutePath(b.rootDir, path)

	if err := b.watcher.Add(path); err != nil {
		return err
	}

	if existing, ok := b.watchers.Get(path); ok {
		b.watchers.Set(path, append(existing, callback))
	} else {
		b.watchers.Set(path, []chan Event{callback})
	}

	return nil
}

func (b *base) Unwatch(path string, callback chan Event) error {
	if b.watchers == nil || b.watcher == nil {
		return nil
	}

	path = AbsolutePath(b.rootDir, path)

	if existing, ok := b.watchers.Get(path); ok {
		newCallbacks := make([]chan Event, 0, len(existing))

		for _, c := range existing {
			if c != callback {
				newCallbacks = append(newCallbacks, c)
			}
		}

		if len(newCallbacks) == 0 {
			b.watchers.Delete(path)
		} else {
			b.watchers.Set(path, newCallbacks)
		}

		if b.watchers.Len() == 0 { // no more watchers
			if err := b.watcher.Close(); err != nil {
				return err
			}
			close(b.closeNotify)
			b.watcher = nil
			b.watchers = nil
		}
	}

	return nil
}

func (b *base) notify() {
	for {
		select {
		case event, isClosed := <-b.watcher.Events:
			if !isClosed {
				return
			}

			list := make([]chan Event, 0)

			listeners, ok := b.watchers.Get(event.Name)
			if !ok {
				b.watchers.Range(func(path string, l []chan Event) bool {
					if strings.HasPrefix(event.Name, path) {
						list = append(list, l...)
					}
					return true
				})
				if len(list) == 0 {
					continue
				}
			} else {
				copy(list, listeners)
			}

			realPath := event.Name
			if b.rootDir != "/" {
				event.Name = "/" + strings.TrimPrefix(event.Name, b.rootDir)
			}

			for _, c := range list {
				var wg sync.WaitGroup
				wg.Add(1)

				go func() {
					defer func() {
						if err := recover(); err != nil {
							// user closed callback channel so we'll remove it from watchers list
							if fmt.Sprintf("%s", err) == "send on closed channel" {
								_ = b.Unwatch(realPath, c)
							}
						}
						wg.Done()
					}()

					if !reflect.ValueOf(c).TrySend(reflect.ValueOf(event)) {
						fmt.Println("failed")
					}
				}()

				wg.Wait()
			}

		case <-b.watcher.Errors:
		case <-b.closeNotify:
			return
		}
	}
}
