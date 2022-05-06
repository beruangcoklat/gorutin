package concurrent

import "sync"

// Execute will process all inputs concurrently by calling the function passed in the arguments.
// The number of goroutines that are used in the concurrent execution could be specified in the numOfRoutines parameter.
// The execution follows fan-out and then fan-in pattern, in which multiple processes are run concurrently, then each
// outputs are gathered and appended to a single slice at the end of the execution. The slice is not guaranteed to have
// the one-to-one order as the input, so it is advised to not rely on the output slice order.
func Execute[TypeIn any, TypeOut any](numOfRoutines int, inputs []TypeIn, process func(input TypeIn) TypeOut) []TypeOut {
	inputChannels := make([](chan TypeIn), numOfRoutines)
	outputChannel := make(chan TypeOut)
	for i := 0; i < numOfRoutines; i++ {
		inputChannels[i] = make(chan TypeIn)
	}

	var wg sync.WaitGroup
	wg.Add(numOfRoutines)

	// spawn workers
	for i := 0; i < numOfRoutines; i++ {
		inputChannel := inputChannels[i]
		go func(inputChan chan TypeIn, outputChan chan TypeOut) {
			defer wg.Done()
			for input := range inputChan {
				outputChan <- process(input)
			}
		}(inputChannel, outputChannel)
	}

	// distribute inputs
	go func(inputs []TypeIn, inputChannels [](chan TypeIn)) {
		for i, input := range inputs {
			channel := i % numOfRoutines
			inputChannels[channel] <- input
		}
		for _, inputChan := range inputChannels {
			close(inputChan)
		}
	}(inputs, inputChannels)

	go func() {
		wg.Wait()
		close(outputChannel)
	}()

	// wait for outputs
	outputs := []TypeOut{}
	for o := range outputChannel {
		outputs = append(outputs, o)
	}

	return outputs
}
