package main

import (
    "github.com/dersebi/golang_exp/exp/inotify"
    "log"
    "strings"
    "time"
)

type Watcher struct {
    Modified chan string
    watch *inotify.Watcher
    watch_list map[string]bool
}

func NewWatcher() (*Watcher, error) {
    watch, err := inotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    w := &Watcher{make(chan string), watch, map[string]bool{}}
    go w.handle()
    return w, nil
}

func (w *Watcher) handle() {
    last := ""
    next := time.Now()
    for {
        select {
        case e := <-w.watch.Error:
            log.Println(e)
        case e := <-w.watch.Event:
            if e.Mask & inotify.IN_MODIFY != 0 {
                t := time.Now()
                if t.Before(next) && e.Name == last {
                    continue
                }
                next = t.Add(1 * time.Second)
                last = e.Name
                w.Modified <- e.Name
            }
        }
    }
}

func (w *Watcher) UpdateWatchList(files [][]byte) error {
    watch_list := map[string]bool{}
    var err error

    for _, b := range files {
        f := strings.TrimSpace(string(b))
        if f != "" {
            log.Println("Watching", f)
            watch_list[f] = true
            if w.watch_list[f] {
                delete(w.watch_list, f)
            } else {
                if err = w.watch.AddWatch(f, inotify.IN_MODIFY); err != nil {
                    log.Println(err)
                }
            }
            watch_list[f] = true
        }
    }

    // unwatch leftovers
    for f := range w.watch_list {
        if err = w.watch.RemoveWatch(f); err != nil {
            log.Println(err)
        }
    }

    w.watch_list = watch_list
    return err
 }
