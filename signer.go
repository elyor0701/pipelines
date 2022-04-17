package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func ExecutePipeline(jobs ...job) {
	fmt.Println("ExecutePipeline :")
	in := make(chan interface{}, 0)
	wg := &sync.WaitGroup{}
	for _, work := range jobs {
		wg.Add(1)
		out := make(chan interface{}, 1)
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
	for data := range in {
		fmt.Println(data, "SingleHash data", data)
		dataMd5 := DataSignerMd5(strconv.Itoa(data.(int)))
		fmt.Println(data, "SingleHash md5(data)", dataMd5)
		dataCrc32Md5 := DataSignerCrc32(dataMd5)
		fmt.Println(data, "SingleHash crc32(md5(data))", dataCrc32Md5)
		dataCrc32 := DataSignerCrc32(strconv.Itoa(data.(int)))
		fmt.Println(data, "SingleHash crc32(data)", dataCrc32)
		result := dataCrc32 + "~" + dataCrc32Md5
		fmt.Println(data, "SingleHash result", result)
		out <- result
		runtime.Gosched()
	}
}

func MultiHash(in, out chan interface{}) {
	for data := range in {
		var result string = ""
		for th := 0; th < 6; th++ {
			addData := strconv.Itoa(th) + data.(string)
			thResult := DataSignerCrc32(addData)
			fmt.Println(data, "MultiHash: crc32(th + step1)", th, thResult)
			result += thResult
		}
		fmt.Println(data, "MultiHash: result", result)
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	data := make([]string, 0)
	for item := range in {
		data = append(data, item.(string))
	}
	sort.Strings(data)
	result := strings.Join(data, "_")
	fmt.Println("CombineResult", result)
	out <- result
}
