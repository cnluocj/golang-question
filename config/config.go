package config

import (
	"encoding/json"
	"golang-question/errorx"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Manager[T any] interface {
	Get() T
	Update(T) errorx.Error
	OnChange(func(T)) (cancel func())
	Watch() Manager[T]     //執行Watch後，會開始監聽配置的變化，並在變化時自動更新 否則每次Get都會從數據源取得最新資料
	InitData(T) Manager[T] //如果數據源沒有資料，則使用InitData put資料
}

type manager[T any] struct {
	config T
}

const configFile = "default_config.json"

func (m manager[T]) Get() T {
	return m.config
}

func IsZeroRef[T any](v T) bool {
	return reflect.ValueOf(&v).Elem().IsZero()
}

func (m *manager[T]) Update(config T) errorx.Error {
	if !IsZeroRef(config) {
		m.config = config
		log.Printf("update: %v\n", m.config)
		return nil
	}
	return m.updateFromLocal()
}

func (m *manager[T]) updateFromLocal() errorx.Error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return errorx.Wrap(err)
	}
	cf, err := os.ReadFile(configPath)
	if err != nil {
		return errorx.Wrap(err)
	}
	if len(cf) == 0 {
		return errorx.New("file empty")
	}
	var dat T
	if err := json.Unmarshal(cf, &dat); err != nil {
		return errorx.Wrap(err)
	}

	m.config = dat
	log.Printf("update: %v\n", m.config)
	return nil
}

func (m manager[T]) OnChange(f func(T)) (cancel func()) {
	log.Println("onChange")
	err := m.updateFromLocal()
	if err != nil {
		log.Println(err)
	}
	return nil
}

func (m *manager[T]) Watch() Manager[T] {
	initWG := sync.WaitGroup{}
	initWG.Add(1)
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Printf("failed to create watch: %s\n", err)
			os.Exit(1)
		}
		defer watcher.Close()

		eventsWG := sync.WaitGroup{}
		eventsWG.Add(1)
		go func() {
			for {
				select {
				case _, ok := <-watcher.Events:
					if !ok {
						eventsWG.Done()
						return
					}
					m.OnChange(nil)
				case err, ok := <-watcher.Errors:
					if ok {
						log.Printf("watch err: %s\n", err)
					}
					eventsWG.Done()
					return
				}
			}
		}()
		configPath, err := getConfigFilePath()
		if err != nil {
			log.Printf("file err: %s\n", err)
		}
		watcher.Add(configPath)
		initWG.Done()
		eventsWG.Wait()
	}()
	initWG.Wait()
	return m
}

func (m *manager[T]) InitData(config T) Manager[T] {
	configPath, err := getConfigFilePath()
	if err != nil {
		log.Printf("file err: %s\n", err)
		m.config = config
		return m
	}
	cf, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("file err: %s\n", err)
		m.config = config
		return m
	}
	if len(cf) == 0 {
		log.Println("file empty")
		m.config = config
		return m
	}
	var dat T
	if err := json.Unmarshal(cf, &dat); err != nil {
		m.config = config
		return m
	}

	m.config = dat
	return m
}

func newMyManager[T any]() *manager[T] {
	return &manager[T]{}
}

func Local[T any]() Manager[T] {
	//TODO: implement
	manager := newMyManager[T]()
	return manager
}

func Etcd[T any]() Manager[T] {
	//TODO: implement
	return nil
}

func getProjectRoot() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	realPath, err := filepath.EvalSymlinks(exPath)
	if err != nil {
		panic(err)
	}
	return realPath, nil
}

func getConfigFilePath() (string, error) {
	root, err := getProjectRoot()
	if err != nil {
		return "", err
	}
	filePath := root + string(os.PathSeparator) + configFile
	log.Println("filepath: " + filePath)
	return filePath, nil

}
