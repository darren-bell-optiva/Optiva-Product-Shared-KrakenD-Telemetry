## KrakenD Elastic Common Schema Logger - Access and Application Logs

The module enables getting GIN router logs in JSON format.

### Setting Up

Clone the [KrakenD-CE](https://github.com/devopsfaith/krakend-ce) repository and update the following changes 

1. In the `router_engine.go` file:

```diff router_engine.go

import (
        "encoding/json"
 
        "github.com/gin-gonic/gin"
+       "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
 
        botdetector "github.com/krakendio/krakend-botdetector/v2/gin"
        httpsecure "github.com/krakendio/krakend-httpsecure/v2/gin"

@@ import (
        opencensus "github.com/krakendio/krakend-opencensus/v2/router/gin"
        "github.com/luraproject/lura/v2/config"
        "github.com/luraproject/lura/v2/core"
+       "github.com/luraproject/lura/v2/proxy"
        luragin "github.com/luraproject/lura/v2/router/gin"
        "github.com/luraproject/lura/v2/transport/http/server"
+
+       optiva_telemetry "github.com/darren-bell-optiva/optiva-product-shared-krakend-telemetry"
 )
 
 // NewEngine creates a new gin engine with some default values and a secure middleware
 @@ func NewEngine(cfg config.ServiceConfig, opt luragin.EngineOptions) *gin.Engine {
        engine := luragin.NewEngine(cfg, opt)
 
+       engine.Use(otelgin.Middleware("krakend"))
+
+       engine.Use(optiva_telemetry.NewGinLogger(cfg.ExtraConfig, gin.LoggerConfig{}))
+
+       jsonWithResponseStatusCodeRender := func(c *gin.Context, response *proxy.Response) {
+               if response == nil {
+                       c.JSON(500, gin.H{})
+                       return
+               }
+               status := response.Metadata.StatusCode
+               c.JSON(status, response.Data)
+       }
+
+       // register the render at the router level
+       luragin.RegisterRender("JsonWithStatusCodeRender", jsonWithResponseStatusCodeRender)
```

2. executor.go

```diff 
@@ import (
 
        "github.com/gin-gonic/gin"
        "github.com/go-contrib/uuid"
        "golang.org/x/sync/errgroup"
 
+       optiva_telemetry "github.com/darren-bell-optiva/optiva-product-shared-krakend-telemetry"
        krakendbf "github.com/krakendio/bloomfilter/v2/krakend"
        asyncamqp "github.com/krakendio/krakend-amqp/v2/async"


@@ type LoggerBuilder struct{}
 // NewLogger sets up the logging components as defined at the configuration.
 func (LoggerBuilder) NewLogger(cfg config.ServiceConfig) (logging.Logger, io.Writer, error) {
        var writers []io.Writer
+
+       telemetryConfig, _ := optiva_telemetry.ConfigGetter(cfg.ExtraConfig)
+       if telemetryConfig != nil {
+               logger, _ := optiva_telemetry.NewApplicationLogger(cfg.ExtraConfig)
+
+               return logger, nil, nil
+       }
+
        gelfWriter, gelfErr := gelf.NewWriter(cfg.ExtraConfig)
        if gelfErr == nil {
                writers = append(writers, gelfWriterWrapper{gelfWriter})

                
@@ func (MetricsAndTraces) Register(ctx context.Context, cfg config.ServiceConfig,
                l.Debug("[SERVICE: OpenCensus] Service correctly registered")
        }
 
+       if err := optiva_telemetry.RegisterOpenTelemetry(ctx, cfg, l); err != nil {
+               if err != optiva_telemetry.ErrNoConfig {
+                       l.Warning("[SERVICE: OpenTelemetry]", err.Error())
+               }
+       } else {
+               l.Debug("[SERVICE: OpenTelemetry] Service correctly registered")
+       }
+
        return metricCollector
 }
```





In KrakenD's `configuration.json` file, add the following to the service `extra_config`:

```json5
  "extra_config": {
    "github_com/darren-bell-optiva/optiva-product-shared-krakend-telemetry": {
        "logging": {
            "level": "INFO",
            "module": "[OPTIVA-TMF-APIGATEWAY]",
            "skip_paths": ["/__health"],
            "json": {
                "disable_html_escape": false,
                "pretty_print": false
            }
            
        },
        "tracing": {
            "exporter_url": "http://{{ env "HOST_IP" }}:14268/api/traces",
            "attributes": {
                "service": "krakend"
            } 
        }

    },
}
```

#### Available Config Options
##### Logging

`level`: The desired log level
`skip_paths`: List of endpoint paths which should not be logged. In the above example configuration, any request to `/__health` will not be logged.
`json.disable_html_escape`: Allows disabling html escaping in output. See https://pkg.go.dev/github.com/sirupsen/logrus#JSONFormatter
`json.pretty_print`: Will indent all json logs. See https://pkg.go.dev/github.com/sirupsen/logrus#JSONFormatter


##### Tracing

`exporter_url` Location to export traces in jaeger format
`attributes` - Additional attributes to apply to the trace

##### Metrics


### TODO - Next steps
Use the default OTEL exported for the traces instead of the Jaeger exporter. e.g. https://docs.honeycomb.io/getting-data-in/opentelemetry/go-distro/