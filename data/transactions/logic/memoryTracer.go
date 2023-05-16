package logic

import (
	"encoding/csv"
	"github.com/algorand/go-algorand/data/basics"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
)

// Debugger is an interface that supports the first version of AVM debuggers.
// It consists of a set of functions called by eval function during AVM program execution.
//
// Deprecated: This interface does not support non-app call or inner transactions. Use EvalTracer
// instead.

type memoryEvalTracerAdaptor struct {
	NullEvalTracer
	debugger    Debugger
	txnDepth    int
	debugState  *DebugState
	opCounter   int
	resolution  int
	csvFileName string
}

func MakeMemoryTracerDebuggerAdaptor(debugger Debugger) EvalTracer {
	return &memoryEvalTracerAdaptor{debugger: debugger,
		opCounter:   0,
		resolution:  100,
		csvFileName: "results.csv",
	}
}

// BeforeTxnGroup updates inner txn depth
func (a *memoryEvalTracerAdaptor) BeforeTxnGroup(ep *EvalParams) {
	a.txnDepth++
}

// AfterTxnGroup updates inner txn depth
func (a *memoryEvalTracerAdaptor) AfterTxnGroup(ep *EvalParams, evalError error) {
	a.txnDepth--
}

// BeforeProgram invokes the debugger's Register hook
func (a *memoryEvalTracerAdaptor) BeforeProgram(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	err := createCSV(a.csvFileName)

	if err != nil {
		log.Printf("an error occured during finalizing the results: %v", err)
	}

	a.debugState = makeMemoryDebugState(cx)
	a.debugger.Register(a.refreshMemoryDebugState(cx, nil, false))
}

// BeforeOpcode invokes the debugger's Update hook
func (a *memoryEvalTracerAdaptor) BeforeOpcode(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Update(a.refreshMemoryDebugState(cx, nil, false))
}

func (a *memoryEvalTracerAdaptor) AfterOpcode(cx *EvalContext, evalError error) {
	if 0 == a.opCounter%a.resolution {
		a.updateResults()
	}
	a.opCounter = a.opCounter + 1
	a.debugger.Update(a.refreshMemoryDebugState(cx, nil, false))
}

// AfterProgram invokes the debugger's Complete hook
func (a *memoryEvalTracerAdaptor) AfterProgram(cx *EvalContext, evalError error) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Complete(a.refreshMemoryDebugState(cx, evalError, true))
}

func makeMemoryDebugState(cx *EvalContext) *DebugState {
	disasm, dsInfo, err := disassembleInstrumented(cx.program, nil)

	if err != nil {
		// Report disassembly error as program text
		disasm = err.Error()
	}

	// initialize DebuggerState with immutable fields
	ds := &DebugState{
		ExecID:      GetProgramID(cx.program),
		Disassembly: disasm,
		PCOffset:    dsInfo.pcOffset,
		GroupIndex:  int(cx.groupIndex),
		TxnGroup:    cx.TxnGroup,
		Proto:       cx.Proto,
	}

	globals := make([]basics.TealValue, len(globalFieldSpecs))
	for _, fs := range globalFieldSpecs {
		// Don't try to grab app only fields when evaluating a signature
		if (cx.runModeFlags&ModeSig) != 0 && fs.mode == ModeApp {
			continue
		}
		sv, err := cx.globalFieldToValue(fs)
		if err != nil {
			sv = stackValue{Bytes: []byte(err.Error())}
		}
		globals[fs.field] = stackValueToTealValue(&sv)
	}
	ds.Globals = globals

	if (cx.runModeFlags & ModeApp) != 0 {
		ds.EvalDelta = cx.txn.EvalDelta
	}

	return ds
}

func (a *memoryEvalTracerAdaptor) updateResults() {
	err := addMemStatsToCSV(a.csvFileName)

	if err != nil {
		log.Printf("an error occured during updating the results: %v", err)
	}
}

func (a *memoryEvalTracerAdaptor) refreshMemoryDebugState(cx *EvalContext, evalError error, finalizeResults bool) *DebugState {
	ds := a.debugState
	// Update pc, line, error, stack, scratch space, callstack,
	// and opcode budget
	ds.PC = cx.pc
	ds.Line = ds.PCToLine(cx.pc)
	if evalError != nil {
		ds.Error = evalError.Error()
	}

	stack := make([]basics.TealValue, len(cx.stack))
	for i, sv := range cx.stack {
		stack[i] = stackValueToTealValue(&sv)
	}

	scratch := make([]basics.TealValue, len(cx.scratch))
	for i, sv := range cx.scratch {
		scratch[i] = stackValueToTealValue(&sv)
	}

	ds.Stack = stack
	ds.Scratch = scratch
	ds.OpcodeBudget = cx.remainingBudget()
	ds.CallStack = ds.parseCallstack(cx.callstack)

	if finalizeResults {
		ds.Disassembly = a.finalizeResults()
	}

	if (cx.runModeFlags & ModeApp) != 0 {
		ds.EvalDelta = cx.txn.EvalDelta
	}

	return ds
}

func (a *memoryEvalTracerAdaptor) finalizeResults() string {
	csvString, err := getCSVAsStringAndDelete(a.csvFileName)
	if err != nil {
		log.Printf("an error occured during finalizing the results: %v", err)
	}

	return csvString
}

func addMemStatsToCSV(filename string) error {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	stats := []string{
		strconv.Itoa(int(mem.HeapAlloc)),
		strconv.Itoa(int(mem.HeapSys)),
		strconv.Itoa(int(mem.HeapIdle)),
		strconv.Itoa(int(mem.HeapInuse)),
		strconv.Itoa(int(mem.StackInuse)),
		strconv.Itoa(int(mem.StackSys)),
	}
	err = writer.Write(stats) // writing stats
	if err != nil {
		return err
	}

	return nil
}

func createCSV(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"heapAlloc", "heapSys", "heapIdle", "heapInuse", "stackInUse", "stackSys"}
	err = writer.Write(headers) // writing header
	if err != nil {
		return err
	}

	return nil
}

func getCSVAsStringAndDelete(filename string) (string, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	err = os.Remove(filename)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}
