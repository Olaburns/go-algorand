//go:build linux
// +build linux

package logic

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/Olaburns/perf-utils"
	"github.com/algorand/go-algorand/data/basics"
	"strconv"
	"time"
)

// Debugger is an interface that supports the first version of AVM debuggers.
// It consists of a set of functions called by eval function during AVM program execution.
//
// Deprecated: This interface does not support non-app call or inner transactions. Use EvalTracer
// instead.

type evalCycleResults struct {
	opcodes []string
	cycles  []string
}

type cycleEvalTracerAdaptor struct {
	NullEvalTracer

	debugger          Debugger
	txnDepth          int
	debugState        *DebugState
	time              time.Time
	elapsedTimeString string
	opcode            OpSpec
	results           *evalCycleResults
	cb                func()
	fd                int
}

func MakeCycleTracerDebuggerAdaptor(debugger Debugger) EvalTracer {
	return &cycleEvalTracerAdaptor{debugger: debugger,
		fd: 0,
		results: &evalCycleResults{
			cycles:  []string{},
			opcodes: []string{},
		}}
}

// BeforeTxnGroup updates inner txn depth
func (a *cycleEvalTracerAdaptor) BeforeTxnGroup(ep *EvalParams) {
	a.txnDepth++
}

// AfterTxnGroup updates inner txn depth
func (a *cycleEvalTracerAdaptor) AfterTxnGroup(ep *EvalParams, evalError error) {
	a.txnDepth--
}

// BeforeProgram invokes the debugger's Register hook
func (a *cycleEvalTracerAdaptor) BeforeProgram(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugState = makeCycleDebugState(cx)
	a.debugger.Register(a.refreshCycleDebugState(cx, nil, false))
}

// BeforeOpcode invokes the debugger's Update hook
func (a *cycleEvalTracerAdaptor) BeforeOpcode(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	opcode := opsByOpcode[LogicVersion][cx.program[cx.pc]]
	a.opcode = opcode
	a.debugger.Update(a.refreshCycleDebugState(cx, nil, false))
	cb, fd, err := perf.StartCPUCycles()
	if err != nil {
		fmt.Println("StartCPUCycles failed:", err)
		_ = writeStringToFile("tracer_log.txt", err.Error())
	}
	writeStringToFile("fd.txt", strconv.Itoa(fd))
	a.cb = cb
	a.fd = fd
}

func (a *cycleEvalTracerAdaptor) AfterOpcode(cx *EvalContext, evalError error) {
	pv, err := perf.StopCPUCycles(a.cb, a.fd)
	if err != nil {
		fmt.Println("StopCPUCycles failed:", err)
		_ = writeStringToFile("tracer_log2.txt", err.Error())
	}
	var cycles int
	if pv != nil {
		cycles = int(pv.Value)
	} else {
		cycles = 0
	}

	a.results.cycles = append(a.results.cycles, strconv.Itoa(cycles))
	a.debugger.Update(a.refreshCycleDebugState(cx, nil, false))
}

// AfterProgram invokes the debugger's Complete hook
func (a *cycleEvalTracerAdaptor) AfterProgram(cx *EvalContext, evalError error) {

	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Complete(a.refreshCycleDebugState(cx, evalError, true))
}

func makeCycleDebugState(cx *EvalContext) *DebugState {
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

func (a *cycleEvalTracerAdaptor) updateResults() {
	a.results.opcodes = append(a.results.opcodes, a.opcode.Name)
	a.results.cycles = append(a.results.cycles, a.elapsedTimeString)
}

func (a *cycleEvalTracerAdaptor) refreshCycleDebugState(cx *EvalContext, evalError error, finalizeResults bool) *DebugState {
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

func (a *cycleEvalTracerAdaptor) finalizeResults() string {
	csvString, err := CycleDataToCSV(a.results.opcodes, a.results.cycles)

	if err != nil {
		fmt.Errorf("an error occured during finalizing the results: %v", err)
	}

	return csvString
}

func CycleDataToCSV(opcodes []string, cycles []string) (string, error) {
	// Check if all slices have the same length
	if len(opcodes) != len(cycles) {
		return "", errors.New("all slices must have the same length")
	}

	// Create a buffer to hold the CSV data
	buf := &bytes.Buffer{}
	w := csv.NewWriter(buf)

	// Write the headers to the CSV
	err := w.Write([]string{"opcodes", "cycles"})
	if err != nil {
		return "", err
	}

	// Write data to CSV
	for i := 0; i < len(opcodes); i++ {
		row := []string{
			opcodes[i],
			cycles[i],
		}
		err = w.Write(row)
		if err != nil {
			return "", err
		}
	}

	// Flush any remaining data to the writer
	w.Flush()

	// Check for any errors during write
	err = w.Error()
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
