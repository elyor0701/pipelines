package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
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
	// inWorker := make(chan ThWorker, 50)
	// outWorker := make(chan ThWorker, 50)
	// for i := 0; i < 20; i++ {
	// 	go func(in, out chan ThWorker) {
	// 		for input := range in {
	// 			dataCrc32 := DataSignerCrc32(input.result)
	// 			out <- ThWorker{input.id, dataCrc32}
	// 		}
	// 	}(inWorker, outWorker)
	// }
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data string) {
			defer wg.Done()
			inWorker := make(chan ThWorker, 5)
			defer close(inWorker)
			outWorker := make(chan ThWorker, 5)
			resultSlice := make([]string, 6)
			for th := 0; th < 6; th++ {
				go func(in, out chan ThWorker) { // should change it
					for input := range in {
						dataCrc32 := DataSignerCrc32(input.result)
						out <- ThWorker{input.id, dataCrc32}
					}
				}(inWorker, outWorker)
				addData := strconv.Itoa(th) + data
				inWorker <- ThWorker{th, addData}
			}
			for i := 0; i < 6; i++ {
				solvedData := <-outWorker
				resultSlice[solvedData.id] = solvedData.result
			}
			result := strings.Join(resultSlice, "")
			out <- result
		}(data.(string))
	}
	wg.Wait()
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
