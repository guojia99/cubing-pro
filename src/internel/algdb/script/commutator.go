package script

import (
	"fmt"
	"os"
	"sync"

	"github.com/2mf8/Better-Bot-Go/log"
	"github.com/dop251/goja"
)

var (
	//vm   *otto.Otto
	vm   *goja.Runtime
	once sync.Once
)

//const CommutatorJs = "./commutator.js"

func InitCommutator(CommutatorJs string) {
	once.Do(func() {
		file, err := os.ReadFile(CommutatorJs)
		if err != nil {
			return
		}

		vm = goja.New()
		_, err = vm.RunString(string(file))
		if err != nil {
			log.Errorf(err.Error())
			vm = nil
			return
		}
		log.Infof("加载js引擎成功")
	})
}

func Commutator(alg string) (string, error) {

	if vm == nil {
		return "-", fmt.Errorf("引擎未启动")
	}
	var commutatorFn func(string) string
	err := vm.ExportTo(vm.Get("commutator"), &commutatorFn)
	if err != nil {
		return "-", err
	}

	out := commutatorFn(alg)
	if out == "Not found." {
		return "-", fmt.Errorf("无")
	}
	return out, nil
}
