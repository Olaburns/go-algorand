package logic

import (
	"github.com/algorand/go-algorand/data/transactions"
	"github.com/algorand/go-algorand/logging"
	"os"
	"strings"
	"unicode"
)

func DetermineTracer(debugger Debugger, stxn *transactions.SignedTxn) EvalTracer {
	if len(stxn.Txn.ApplicationArgs) > 1 {
		tracerType := string(stxn.Txn.ApplicationArgs[1])
		tracerType = removeLeadingChars(tracerType)
		logging.Base().Info("Test: Arrived at Determine Tracer")
		switch tracerType {
		case "timingTracer":
			return MakeTimingTracerDebuggerAdaptor(debugger)
		case "memoryTracer":
			return MakeMemoryTracerDebuggerAdaptor(debugger)
		case "memoryTransactionTracer":
			return MakeMemoryTransactionTracerDebuggerAdaptor(debugger)
		default:
			return MakeEvalTracerDebuggerAdaptor(debugger)
		}
	} else {
		return MakeEvalTracerDebuggerAdaptor(debugger)
	}
}

func writeStringToFile(filename, data string) error {
	// Open the file for writing, creating it if it doesn't exist
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write the data to the file
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	return nil
}

func removeLeadingChars(str string) string {
	/*
		var result string
		for _, ch := range str {
			if unicode.IsLetter(ch) {
				result += string(ch)
				break
			}
		}
	*/
	return str[strings.IndexFunc(str, unicode.IsLetter):]
}
