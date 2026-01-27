File: server
Build ID: d8d18c8780372a1ad258a6a54971386f99305d09
Type: inuse_space
Time: 2026-01-11 16:46:56 MSK
Showing nodes accounting for -22.28MB, 43.54% of 51.17MB total
Dropped 1 node (cum <= 0.26MB)
      flat  flat%   sum%        cum   cum%
  -22.92MB 44.79% 44.79%   -24.53MB 47.94%  compress/flate.NewWriter (inline)
   -1.61MB  3.15% 47.94%    -1.61MB  3.15%  compress/flate.(*compressor).initDeflate (inline)
    1.50MB  2.94% 45.00%     1.50MB  2.94%  runtime.allocm
   -0.75MB  1.47% 46.47%    -0.75MB  1.47%  go.uber.org/zap/zapcore.newCounters (inline)
   -0.50MB  0.98% 47.45%    -0.50MB  0.98%  bufio.NewReaderSize (inline)
    0.50MB  0.98% 46.47%     0.50MB  0.98%  runtime.gcBgMarkWorker
    0.50MB  0.98% 45.49%     0.50MB  0.98%  runtime.malg
    0.50MB  0.98% 44.51%     0.50MB  0.98%  net/http.readRequest
    0.50MB  0.98% 43.54%     0.50MB  0.98%  net/http.(*Server).newConn (inline)
   -0.50MB  0.98% 44.51%    -0.50MB  0.98%  compress/flate.(*huffmanEncoder).generate
   -0.50MB  0.98% 45.49%    -0.50MB  0.98%  runtime.acquireSudog
    0.50MB  0.98% 44.51%     0.50MB  0.98%  bufio.NewWriterSize (inline)
    0.50MB  0.98% 43.54%     0.50MB  0.98%  syscall.anyToSockaddr
         0     0% 43.54%    -0.50MB  0.98%  bufio.NewReader (inline)
         0     0% 43.54%    -0.50MB  0.98%  compress/flate.(*Writer).Close (inline)
         0     0% 43.54%    -0.50MB  0.98%  compress/flate.(*compressor).close
         0     0% 43.54%    -0.50MB  0.98%  compress/flate.(*compressor).deflate
         0     0% 43.54%    -1.61MB  3.15%  compress/flate.(*compressor).init
         0     0% 43.54%    -0.50MB  0.98%  compress/flate.(*compressor).writeBlock
         0     0% 43.54%    -0.50MB  0.98%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 43.54%    -0.50MB  0.98%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 43.54%    -0.50MB  0.98%  compress/gzip.(*Writer).Close
         0     0% 43.54%   -24.53MB 47.94%  compress/gzip.(*Writer).Write
         0     0% 43.54%   -25.03MB 48.91%  github.com/go-chi/chi/v5.(*ChainHandler).ServeHTTP
         0     0% 43.54%   -25.03MB 48.91%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 43.54%   -25.03MB 48.91%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 43.54%   -24.53MB 47.94%  github.com/htrandev/metrics/internal/handler.(*MetricHandler).GetAll
         0     0% 43.54%    -0.50MB  0.98%  github.com/htrandev/metrics/internal/handler/middleware.(*compressWriter).Close
         0     0% 43.54%   -24.53MB 47.94%  github.com/htrandev/metrics/internal/handler/middleware.(*compressWriter).Write
         0     0% 43.54%   -43.91MB 85.81%  github.com/htrandev/metrics/internal/router.New.Compress.func4.1
         0     0% 43.54%    18.88MB 36.89%  github.com/htrandev/metrics/internal/router.New.Compress.func6.1
         0     0% 43.54%   -43.91MB 85.81%  github.com/htrandev/metrics/internal/router.New.Logger.func2.1
         0     0% 43.54%    18.88MB 36.89%  github.com/htrandev/metrics/internal/router.New.Logger.func3.1
         0     0% 43.54%   -25.03MB 48.91%  github.com/htrandev/metrics/internal/router.New.MethodChecker.func1.1
         0     0% 43.54%   -43.91MB 85.81%  github.com/htrandev/metrics/internal/router.New.Sign.func3.1
         0     0% 43.54%    18.88MB 36.89%  github.com/htrandev/metrics/internal/router.New.Sign.func5.1
         0     0% 43.54%    -0.75MB  1.47%  github.com/htrandev/metrics/pkg/logger.NewZapLogger
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap.(*Logger).WithOptions
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap.Config.Build
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap.Config.buildOptions.WrapCore.func5
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap.Config.buildOptions.func1
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap.New
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap.optionFunc.apply
         0     0% 43.54%    -0.75MB  1.47%  go.uber.org/zap/zapcore.NewSamplerWithOptions
         0     0% 43.54%     0.50MB  0.98%  internal/poll.(*FD).Accept
         0     0% 43.54%     0.50MB  0.98%  internal/poll.accept
         0     0% 43.54%    -0.75MB  1.47%  main.main
         0     0% 43.54%    -0.75MB  1.47%  main.run
         0     0% 43.54%        1MB  1.95%  main.run.func2
         0     0% 43.54%     0.50MB  0.98%  net.(*TCPListener).Accept
         0     0% 43.54%     0.50MB  0.98%  net.(*TCPListener).accept
         0     0% 43.54%     0.50MB  0.98%  net.(*netFD).accept
         0     0% 43.54%        1MB  1.95%  net/http.(*Server).ListenAndServe
         0     0% 43.54%        1MB  1.95%  net/http.(*Server).Serve
         0     0% 43.54%   -25.03MB 48.92%  net/http.(*conn).serve
         0     0% 43.54%   -25.03MB 48.91%  net/http.HandlerFunc.ServeHTTP
         0     0% 43.54%    -0.50MB  0.98%  net/http.newBufioReader
         0     0% 43.54%     0.50MB  0.98%  net/http.newBufioWriterSize
         0     0% 43.54%   -25.03MB 48.91%  net/http.serverHandler.ServeHTTP
         0     0% 43.54%     0.50MB  0.98%  runtime.(*scavengerState).wake
         0     0% 43.54%    -0.50MB  0.98%  runtime.ensureSigM.func1
         0     0% 43.54%    -0.50MB  0.98%  runtime.gopreempt_m (inline)
         0     0% 43.54%    -0.50MB  0.98%  runtime.goschedImpl
         0     0% 43.54%     0.50MB  0.98%  runtime.injectglist
         0     0% 43.54%     0.50MB  0.98%  runtime.injectglist.func1
         0     0% 43.54%    -0.75MB  1.47%  runtime.main
         0     0% 43.54%     0.50MB  0.98%  runtime.mcall
         0     0% 43.54%    -0.50MB  0.98%  runtime.morestack
         0     0% 43.54%     1.50MB  2.94%  runtime.mstart
         0     0% 43.54%     1.50MB  2.94%  runtime.mstart0
         0     0% 43.54%     1.50MB  2.94%  runtime.mstart1
         0     0% 43.54%     1.50MB  2.94%  runtime.newm
         0     0% 43.54%     0.50MB  0.98%  runtime.newproc.func1
         0     0% 43.54%     0.50MB  0.98%  runtime.newproc1
         0     0% 43.54%    -0.50MB  0.98%  runtime.newstack
         0     0% 43.54%     0.50MB  0.98%  runtime.park_m
         0     0% 43.54%        1MB  1.96%  runtime.resetspinning
         0     0% 43.54%        1MB  1.96%  runtime.schedule
         0     0% 43.54%    -0.50MB  0.98%  runtime.selectgo
         0     0% 43.54%     1.50MB  2.94%  runtime.startm
         0     0% 43.54%     0.50MB  0.98%  runtime.sysmon
         0     0% 43.54%     0.50MB  0.98%  runtime.systemstack
         0     0% 43.54%        1MB  1.96%  runtime.wakep
         0     0% 43.54%     0.50MB  0.98%  syscall.Accept4