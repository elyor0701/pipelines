package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func ExecutePipeline(jobs ...job) {
	fmt.Println("ExecutePipeline :")
	in := make(chan interface{}, 0)
	wg := &sync.WaitGroup{}
	for _, work := range jobs {
		wg.Add(1)
		out := make(chan interface{}, 100)
		go func(in, out chan interface{}, work job, wg *sync.WaitGroup) {
			defer wg.Done()
			work(in, out)
			close(out)
		}(in, out, work, wg)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	start := time.Now()
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	for data := range in {
		wg.Add(1)
		go func(data string) { // check without chan parametr
			defer wg.Done()
			dataCrc32Ch := SingleCrc32Helper(data)
			mu.Lock()
			dataMd5 := DataSignerMd5(data)
			mu.Unlock()
			dataCrc32Md5 := DataSignerCrc32(dataMd5)
			dataCrc32 := <-dataCrc32Ch
			result := dataCrc32 + "~" + dataCrc32Md5
			out <- result
		}(strconv.Itoa(data.(int)))
	}
	wg.Wait()
	end := time.Since(start)
	fmt.Println("SingleH", end)
}

// func SingleHash(in, out chan interface{}) {
// 	start := time.Now()
// 	for data := range in {
// 		dataCrc32Ch := SingleCrc32Helper(strconv.Itoa(data.(int)))
// 		dataMd5 := DataSignerMd5(strconv.Itoa(data.(int)))
// 		dataCrc32Md5 := DataSignerCrc32(dataMd5)
// 		dataCrc32 := <-dataCrc32Ch
// 		result := dataCrc32 + "~" + dataCrc32Md5
// 		out <- result
// 	}
// 	end := time.Since(start)
// 	fmt.Println("Single", end)
// }

func SingleCrc32Helper(data string) chan string {
	result := make(chan string, 1)
	go func(out chan<- string) {
		dataCrc32 := DataSignerCrc32(data)
		out <- dataCrc32
	}(result)
	return result
}

// func MultiHash(in, out chan interface{}) {
// 	for data := range in {
// 		var result string = ""
// 		for th := 0; th < 6; th++ {
// 			addData := strconv.Itoa(th) + data.(string)
// 			thResult := DataSignerCrc32(addData)
// 			result += thResult
// 		}
// 		out <- result
// 	}
// }

type ThWorker struct {
	id     int
	result string
}

func MultiHash(in, out chan interface{}) {
	inWorker := make(chan ThWorker, 5)
	outWorker := make(chan ThWorker, 5)
	for i := 0; i < 6; i++ {
		go func(in, out chan ThWorker) {
			for input := range in {
				dataCrc32 := DataSignerCrc32(input.result)
				out <- ThWorker{input.id, dataCrc32}
			}
		}(inWorker, outWorker)
	}
	for data := range in {
		resultSlice := make([]string, 6)
		for th := 0; th < 6; th++ {
			addData := strconv.Itoa(th) + data.(string)
			inWorker <- ThWorker{th, addData}
		}
		for i := 0; i < 6; i++ {
			solvedData := <-outWorker
			resultSlice[solvedData.id] = solvedData.result
		}
		result := strings.Join(resultSlice, "")
		out <- result
	}
	close(inWorker)
}

func CombineResults(in, out chan interface{}) {
	data := make([]string, 0)
	for item := range in {
		data = append(data, item.(string))
	}
	sort.Strings(data)
	result := strings.Join(data, "_")
	out <- result
}
