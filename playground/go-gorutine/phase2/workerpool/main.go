package main

import (
	"fmt"
	"sync"
)

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("worker %d processing job %d\n", id, job)
	}
}

func main() {
	jobs := make(chan int)
	var wg sync.WaitGroup

	workerCount := 3

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	for job := 1; job <= 10; job++ {
		jobs <- job
	}

	close(jobs)
	wg.Wait()
}
