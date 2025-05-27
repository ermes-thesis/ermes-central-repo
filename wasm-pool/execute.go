package wasm

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (exec *WasmExecutor) InvokeFn(fnName string) {}

func (exec *WasmExecutor) InvokeAuthFn(fnName string, req *http.Request) {
	vm := exec.vm
	err := vm.LoadWasmFile(fnName)
	if err != nil {
		log.Fatalf("Failed to load wasm file: %v", err)
	}

	err = vm.Validate()
	if err != nil {
		log.Fatalf("Failed to validate wasm file: %v", err)
	}

	wasmInput, err := prepareWASMInput(req)
	lengthWasmInput := len(wasmInput)

	mallocRes, err := vm.Execute("malloc", int32(lengthWasmInput))
	if err != nil {
		log.Fatalf("Failed to execute malloc: %v", err)
	}
	wasmInputMemPtr := mallocRes[0].(int32)

	mod := vm.GetActiveModule()
	mem := mod.FindMemory("memory")
	err = mem.SetData(wasmInput, uint(wasmInputMemPtr), uint(lengthWasmInput))
	if err != nil {
		log.Fatalf("Failed to set memory: %v", err)
	}

	//authFnRes, err := vm.Execute("entry", wasmInputMemPtr, int32(lengthWasmInput))
	//authFnResPtr := authFnRes[0].(int32)

	//memData, err := mem.GetData(uint(authFnResPtr), 8)

}

type AuthFnInput struct {
	Method      string              `json:"method"`
	URL         string              `json:"url"`
	Header      map[string][]string `json:"header"`
	Body        string              `json:"body"`
	RemoteAddr  string              `json:"remoteAddr"`
	QueryParams map[string][]string `json:"queryParams"`
}

func prepareWASMInput(r *http.Request) ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	reqInput := AuthFnInput{
		Method:      r.Method,
		URL:         r.URL.String(),
		Header:      map[string][]string(r.Header),
		Body:        string(bodyBytes),
		RemoteAddr:  r.RemoteAddr,
		QueryParams: r.URL.Query(),
	}
	return json.Marshal(reqInput)
}
