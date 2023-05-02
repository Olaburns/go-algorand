package logic

/*
import (
	"github.com/algorand/go-algorand/data/basics"
	"strconv"
	"time"
)

package logic

import (
"strconv"
"time"

"github.com/algorand/go-algorand/data/basics"
)

// Debugger is an interface that supports the first version of AVM debuggers.
// It consists of a set of functions called by eval function during AVM program execution.
//
// Deprecated: This interface does not support non-app call or inner transactions. Use EvalTracer
// instead.

type memoryEvalTracerAdaptor struct {
	NullEvalTracer

	debugger          Debugger
	txnDepth          int
	debugState        *DebugState
	time              time.Time
	elapsedTimeString string
	opcode            OpSpec
}

func MakeMemoryTracerDebuggerAdaptor(debugger Debugger) EvalTracer {
	return &memoryEvalTracerAdaptor{debugger: debugger}
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
	a.debugState = makeTimingDebugState(cx)
	a.debugger.Register(a.refreshMemoryDebugState(cx, nil, false))
}

// BeforeOpcode invokes the debugger's Update hook
func (a *memoryEvalTracerAdaptor) BeforeOpcode(cx *EvalContext) {
	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	opcode := opsByOpcode[LogicVersion][cx.program[cx.pc]]
	a.opcode = opcode
	a.debugger.Update(a.refreshMemoryDebugState(cx, nil, false))
	a.time = time.Now()
}

func (a *memoryEvalTracerAdaptor) AfterOpcode(cx *EvalContext, evalError error) {
	elapsedTime := strconv.FormatInt(time.Since(a.time).Nanoseconds(), 10)
	a.elapsedTimeString = elapsedTime
	a.debugger.Update(a.refreshMemoryDebugState(cx, nil, true))
}

// AfterProgram invokes the debugger's Complete hook
func (a *memoryEvalTracerAdaptor) AfterProgram(cx *EvalContext, evalError error) {

	if a.txnDepth > 0 {
		// only report updates for top-level transactions, for backwards compatibility
		return
	}
	a.debugger.Complete(a.refreshMemoryDebugState(cx, evalError, false))
}

func makeMemoryDebugState(cx *EvalContext) *DebugState {
	_, dsInfo, err := disassembleInstrumented(cx.program, nil)

	// Disasm is just a placeholder for the observed results
	disasm := "Opcode, Timing\n"
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

func (a *memoryEvalTracerAdaptor) updateMemoryString(ds *DebugState) string {

	result := ds.Disassembly + a.opcode.Name + "," + a.elapsedTimeString + "\n"
	writeStringToFile("UpdatetTimingString.txt", result)
	return result
}

func (a *memoryEvalTracerAdaptor) refreshMemoryDebugState(cx *EvalContext, evalError error, addMemoryStats bool) *DebugState {
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
	if addMemoryStats {
		ds.Disassembly = a.updateMemoryString(ds)
	}

	if (cx.runModeFlags & ModeApp) != 0 {
		ds.EvalDelta = cx.txn.EvalDelta
	}

	return ds
}
*/
