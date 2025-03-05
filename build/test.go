package main

import (
	"fmt"
	"math"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var highLoadActive bool
var wg sync.WaitGroup

// High load function
type HighLoadTask struct {
	stop chan bool
}

func NewHighLoadTask() *HighLoadTask {
	return &HighLoadTask{stop: make(chan bool)}
}

func (h *HighLoadTask) Start() {
	cpuCount := runtime.NumCPU()
	runtime.GOMAXPROCS(cpuCount)
	fmt.Println("Starting high load on", cpuCount, "CPU cores")

	for i := 0; i < cpuCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-h.stop:
					return
				default:
					// Perform CPU-intensive task
					_ = math.Sqrt(float64(time.Now().UnixNano()))
				}
			}
		}()
	}
}

func (h *HighLoadTask) Stop() {
	close(h.stop)
	wg.Wait()
	fmt.Println("High load stopped")
}

func main() {
	r := gin.Default()
	var highLoad *HighLoadTask

	r.GET("/start-high-load", func(c *gin.Context) {
		if !highLoadActive {
			highLoad = NewHighLoadTask()
			highLoad.Start()
			highLoadActive = true
			c.JSON(http.StatusOK, gin.H{"message": "High load started"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "High load already running"})
		}
	})

	r.GET("/stop-high-load", func(c *gin.Context) {
		if highLoadActive {
			highLoad.Stop()
			highLoadActive = false
			c.JSON(http.StatusOK, gin.H{"message": "High load stopped"})
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "High load not running"})
		}
	})

	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"highLoadActive": highLoadActive})
	})

	r.Run(":8080")
}
