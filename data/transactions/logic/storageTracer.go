package logic

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/algorand/go-algorand/data/basics"
	"os"
	"strconv"
	"strings"
)

// Debugger is an interface that supports the first version of AVM debuggers.
// It consists of a set of functions called by eval function during AVM program execution.
//
// Deprecated: This interface does not support non-app call or inner transactions. Use EvalTracer
// instead.

type ProcIO struct {
	Rchar               int64
	Wchar               int64
	Syscr               int64
	Syscw               int64
	ReadBytes           int64
	WriteBytes          int64
	CancelledWriteBytes int64
}

type storageEvalTracerAdaptor struct {
	NullEvalTracer

	debugger   Debugger
	txnDepth   int
	debugState *DebugState
	results    []*ProcIO
	resolution int
	opCounter  int
}

func MakeStroageTracerDebuggerAdaptor(debugger Debugger) EvalTracer {
	return &storageEvalTracerAdaptor{
		debugger:   debugger,
		results:    []*ProcIO{},
		resolution: 100,
		opCounter:  0,
	}
}

// BeforeTxnGroup updates inner txn depth
func (a *storageEvalTracerAdaptor) BeforeTxnGroup(ep *EvalParams) {
	a.txnDepth++
}

// AfterTxnGroup updates inner txn depth
func (a *storageEvalTracerAdaptor) AfterTxnGroup(ep *EvalParams, evalError error) {
	a.txnDepth--
}

// BeforeProgram invokes the debugger's Register hook
func (a *storageEvalTracerAdaptor) BeforeProgram(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.updateResults()
	a.debugState = makeStorageDebugState(cx)
	a.debugger.Register(a.refreshStroageDebugState(cx, nil, false))
}

// BeforeOpcode invokes the debugger's Update hook
func (a *storageEvalTracerAdaptor) BeforeOpcode(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Update(a.refreshStroageDebugState(cx, nil, false))
}

func (a *storageEvalTracerAdaptor) AfterOpcode(cx *EvalContext, evalError error) {
	if 0 == a.opCounter%a.resolution {
		a.updateResults()
	}
	a.debugger.Update(a.refreshStroageDebugState(cx, nil, false))
	a.opCounter = a.opCounter + 1
}

// AfterProgram invokes the debugger's Complete hook
func (a *storageEvalTracerAdaptor) AfterProgram(cx *EvalContext, evalError error) {

	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.updateResults()
	a.debugger.Complete(a.refreshStroageDebugState(cx, evalError, true))
}

func makeStorageDebugState(cx *EvalContext) *DebugState {
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

func (a *storageEvalTracerAdaptor) updateResults() {
	a.readProcessStats()
}

func (a *storageEvalTracerAdaptor) refreshStroageDebugState(cx *EvalContext, evalError error, finalizeResults bool) *DebugState {
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

func (a *storageEvalTracerAdaptor) finalizeResults() string {
	csvString, err := procIOToCSV(a.results)

	if err != nil {
		fmt.Errorf("an error occured during finalizing the results: %v", err)
	}

	return csvString
}

func (a *storageEvalTracerAdaptor) readProcessStats() {
	pid := os.Getpid()
	pidStr := strconv.Itoa(pid)
	pMetrics, err := ReadProcIO(pidStr)
	if err != nil {
		fmt.Errorf("Can not read metrics %v", err)
	}
	a.results = append(a.results, pMetrics)
}

func ReadProcIO(pid string) (*ProcIO, error) {
	file, err := os.Open(fmt.Sprintf("/proc/%s/io", pid))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	result := &ProcIO{}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ": ")
		if len(parts) != 2 {
			continue
		}

		value, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}

		switch parts[0] {
		case "rchar":
			result.Rchar = value
		case "wchar":
			result.Wchar = value
		case "syscr":
			result.Syscr = value
		case "syscw":
			result.Syscw = value
		case "read_bytes":
			result.ReadBytes = value
		case "write_bytes":
			result.WriteBytes = value
		case "cancelled_write_bytes":
			result.CancelledWriteBytes = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func procIOToCSV(procIOs []*ProcIO) (string, error) {
	// Create a buffer to write our output to
	b := &bytes.Buffer{}

	// Create a CSV writer that writes to our buffer
	writer := csv.NewWriter(b)

	// Write the header to the CSV file
	if err := writer.Write([]string{"Rchar", "Wchar", "Syscr", "Syscw", "ReadBytes", "WriteBytes"}); err != nil {
		return "", err
	}

	// Iterate through the input and write each ProcIO's data to the CSV writer
	for _, procIO := range procIOs {
		record := []string{
			strconv.FormatInt(procIO.Rchar, 10),
			strconv.FormatInt(procIO.Wchar, 10),
			strconv.FormatInt(procIO.Syscr, 10),
			strconv.FormatInt(procIO.Syscw, 10),
			strconv.FormatInt(procIO.ReadBytes, 10),
			strconv.FormatInt(procIO.WriteBytes, 10),
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	// Flush any remaining data from the writer to the buffer
	writer.Flush()

	// Check for any error that occurred during the write
	if err := writer.Error(); err != nil {
		return "", err
	}

	// Convert the buffer's contents to a string and return it
	return b.String(), nil
}
