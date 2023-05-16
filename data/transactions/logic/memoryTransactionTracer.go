package logic

import (
	"fmt"
	"github.com/algorand/go-algorand/data/basics"
)

// Debugger is an interface that supports the first version of AVM debuggers.
// It consists of a set of functions called by eval function during AVM program execution.
//
// Deprecated: This interface does not support non-app call or inner transactions. Use EvalTracer
// instead.

type memoryTransactionEvalTracerAdaptor struct {
	NullEvalTracer
	debugger    Debugger
	txnDepth    int
	debugState  *DebugState
	csvFileName string
}

func MakeMemoryTransactionTracerDebuggerAdaptor(debugger Debugger) EvalTracer {
	return &memoryTransactionEvalTracerAdaptor{debugger: debugger}
}

// BeforeTxnGroup updates inner txn depth
func (a *memoryTransactionEvalTracerAdaptor) BeforeTxnGroup(ep *EvalParams) {
	a.txnDepth++
}

// AfterTxnGroup updates inner txn depth
func (a *memoryTransactionEvalTracerAdaptor) AfterTxnGroup(ep *EvalParams, evalError error) {
	a.txnDepth--
}

// BeforeProgram invokes the debugger's Register hook
func (a *memoryTransactionEvalTracerAdaptor) BeforeProgram(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}

	err := createCSV(a.csvFileName)

	if err != nil {
		fmt.Errorf("an error occured during finalizing the results: %v", err)
	}
	a.debugState = makeMemoryTransactionDebugState(cx)
	a.debugger.Register(a.refreshMemoryTransactionDebugState(cx, nil, false))
	a.updateResults()
}

// BeforeOpcode invokes the debugger's Update hook
func (a *memoryTransactionEvalTracerAdaptor) BeforeOpcode(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Update(a.refreshMemoryTransactionDebugState(cx, nil, false))
}

func (a *memoryTransactionEvalTracerAdaptor) AfterOpcode(cx *EvalContext, evalError error) {
	a.debugger.Update(a.refreshMemoryTransactionDebugState(cx, nil, false))
}

// AfterProgram invokes the debugger's Complete hook
func (a *memoryTransactionEvalTracerAdaptor) AfterProgram(cx *EvalContext, evalError error) {
	a.updateResults()
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Complete(a.refreshMemoryTransactionDebugState(cx, evalError, true))
}

func makeMemoryTransactionDebugState(cx *EvalContext) *DebugState {
	_, dsInfo, err := disassembleInstrumented(cx.program, nil)

	// Disasm is just a placeholder for the observed results
	disasm := ""
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

func (a *memoryTransactionEvalTracerAdaptor) updateResults() {
	err := addMemStatsToCSV(a.csvFileName)

	if err != nil {
		fmt.Errorf("an error occured during updating the results: %v", err)
	}
}

func (a *memoryTransactionEvalTracerAdaptor) refreshMemoryTransactionDebugState(cx *EvalContext, evalError error, finalizeResults bool) *DebugState {
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

func (a *memoryTransactionEvalTracerAdaptor) finalizeResults() string {
	csvString, err := getCSVAsStringAndDelete(a.csvFileName)

	if err != nil {
		fmt.Errorf("an error occured during finalizing the results: %v", err)
	}

	return csvString
}
