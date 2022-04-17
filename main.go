package main

import "fmt"

func main() {
	var testResult string
	SignJobs := []job{
		job(func(in, out chan interface{}) {
			inputData := []int{0, 1}
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			testResult = data
			fmt.Println("testResult :", testResult)
		}),
	}
	fmt.Println("Main :")
	ExecutePipeline(SignJobs...)
	fmt.Println("testResult in Main :", testResult)
}
