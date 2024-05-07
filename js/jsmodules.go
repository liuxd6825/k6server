package js

import (
	"errors"
	"sync"

	"github.com/liuxd6825/k6server/ext"
	"github.com/liuxd6825/k6server/js/common"
	"github.com/liuxd6825/k6server/js/modules"
	"github.com/liuxd6825/k6server/js/modules/k6"
	"github.com/liuxd6825/k6server/js/modules/k6/crypto"
	"github.com/liuxd6825/k6server/js/modules/k6/crypto/x509"
	"github.com/liuxd6825/k6server/js/modules/k6/data"
	"github.com/liuxd6825/k6server/js/modules/k6/encoding"
	"github.com/liuxd6825/k6server/js/modules/k6/execution"
	"github.com/liuxd6825/k6server/js/modules/k6/experimental/fs"
	"github.com/liuxd6825/k6server/js/modules/k6/experimental/streams"
	"github.com/liuxd6825/k6server/js/modules/k6/experimental/tracing"
	"github.com/liuxd6825/k6server/js/modules/k6/grpc"
	"github.com/liuxd6825/k6server/js/modules/k6/html"
	"github.com/liuxd6825/k6server/js/modules/k6/http"
	"github.com/liuxd6825/k6server/js/modules/k6/metrics"
	"github.com/liuxd6825/k6server/js/modules/k6/timers"
	"github.com/liuxd6825/k6server/js/modules/k6/ws"

	"github.com/grafana/xk6-browser/browser"
	"github.com/grafana/xk6-redis/redis"
	"github.com/grafana/xk6-webcrypto/webcrypto"
	expws "github.com/grafana/xk6-websockets/websockets"
)

func getInternalJSModules() map[string]interface{} {
	return map[string]interface{}{
		"k6":                         k6.New(),
		"k6/crypto":                  crypto.New(),
		"k6/crypto/x509":             x509.New(),
		"k6/data":                    data.New(),
		"k6/encoding":                encoding.New(),
		"k6/timers":                  timers.New(),
		"k6/execution":               execution.New(),
		"k6/experimental/redis":      redis.New(),
		"k6/experimental/streams":    streams.New(),
		"k6/experimental/webcrypto":  webcrypto.New(),
		"k6/experimental/websockets": &expws.RootModule{},
		"k6/experimental/timers": newWarnExperimentalModule(timers.New(),
			"k6/experimental/timers is now part of the k6 core, please change your imports to use k6/timers instead."+
				" The k6/experimental/timers will be removed in k6 v0.52.0"),
		"k6/experimental/tracing": tracing.New(),
		"k6/experimental/browser": browser.New(),
		"k6/experimental/fs":      fs.New(),
		"k6/net/grpc":             grpc.New(),
		"k6/html":                 html.New(),
		"k6/http":                 http.New(),
		"k6/metrics":              metrics.New(),
		"k6/ws":                   ws.New(),
		"k6/experimental/grpc": newRemovedModule(
			"k6/experimental/grpc has been graduated, please use k6/net/grpc instead." +
				" See https://grafana.com/docs/k6/latest/javascript-api/k6-net-grpc/ for more information.",
		),
	}
}

func getJSModules() map[string]interface{} {
	result := getInternalJSModules()
	external := ext.Get(ext.JSExtension)

	// external is always prefixed with `k6/x`
	for _, e := range external {
		result[e.Name] = e.Module
	}

	return result
}

type warnExperimentalModule struct {
	once *sync.Once
	msg  string
	base modules.Module
}

func newWarnExperimentalModule(base modules.Module, msg string) modules.Module {
	return &warnExperimentalModule{
		msg:  msg,
		base: base,
		once: &sync.Once{},
	}
}

func (w *warnExperimentalModule) NewModuleInstance(vu modules.VU) modules.Instance {
	w.once.Do(func() { vu.InitEnv().Logger.Warn(w.msg) })
	return w.base.NewModuleInstance(vu)
}

type removedModule struct {
	errMsg string
}

func newRemovedModule(errMsg string) modules.Module {
	return &removedModule{errMsg: errMsg}
}

func (rm *removedModule) NewModuleInstance(vu modules.VU) modules.Instance {
	common.Throw(vu.Runtime(), errors.New(rm.errMsg))

	return nil
}
