## Envoy External Processing filter for decoding Google gRPC PubSub Messages

Just an envoy proxy that decodes and prints google pubsub messages.

Basically, another variation of these article

- [Decoding gRPC Messages using Envoy](https://blog.salrashid.dev/articles/2022/envoy_grpc_decode/)
- [Envoy Dynamic Forward Proxy configuration with Downstream SNI for Google APIs and httpbin](https://blog.salrashid.dev/articles/2022/envoy_dynamic_forward_proxy_with_sni/)

![images/pubsub_ext_proxy.png](images/pubsub_ext_proxy.png)


At the moment, the proxy just decodes the messages and sends the original message as-is upstream to gcp.

if you want, you can alter the data code to optionally encrypt some fields or add payload metadata to each request or response.  Or even filter based on the topic the messages are intended for.

The current filter just inspects pubsub messages but your'e free to alter it for any Google GCP gRPC service or even easier, HTTP/Rest service

The setup is easy

```bash
# create a topic to send messages to 
gcloud pubsub topics create topic1
gcloud pubsub subscriptions create ff1-subscribe --topic=topic1
```

Then run the filter

```bash
cd ext_proc/
go run filter.go
```

Run envoy
```bash
docker cp `docker create envoyproxy/envoy-dev:latest`:/usr/local/bin/envoy /tmp/
/tmp/envoy -c envoy_server.yaml -l debug
```

Run  pubsub client
```bash
cd pubsub_client/
go run main.go --projectID $PROJECT_ID
```



---

### Pubsub Client

```bash
$ go run main.go 
Published message msg ID: 6016065317030901
Published message msg ID: 6016065317030900
```

#### External Processor logs

```log
$ go run filter.go 
2022/10/21 17:00:36 Starting server...
2022/10/21 17:00:38 Handling grpc Check request + service:"envoy.service.ext_proc.v3.ExternalProcessor"
2022/10/21 17:00:40 Got stream:  -->  
2022/10/21 17:00:40 pb.ProcessingRequest_RequestHeaders &{headers:{headers:{key:":method"  value:"POST"}  headers:{key:":scheme"  value:"https"}  headers:{key:":path"  value:"/google.pubsub.v1.Publisher/Publish"}  headers:{key:":authority"  value:"pubsub.googleapis.com"}  headers:{key:"content-type"  value:"application/grpc"}  headers:{key:"user-agent"  value:"grpc-go/1.48.0"}  headers:{key:"te"  value:"trailers"}  headers:{key:"grpc-timeout"  value:"59870023u"}  headers:{key:"authorization"  value:"Bearer ya29.redacted"}  headers:{key:"grpc-tags-bin"  value:"AAAGc3RhdHVzAk9LAAV0b3BpYyhwcm9qZWN0cy9mYWJsZWQtcmF5LTEwNDExNy90b3BpY3MvdG9waWMx"}  headers:{key:"x-goog-api-client"  value:"gl-go/1.19.0 gccl/1.25.1 gapic/1.25.1 gax/2.4.0 grpc/1.48.0"}  headers:{key:"x-goog-request-params"  value:"topic=projects%2Ffabled-ray-104117%2Ftopics%2Ftopic1"}  headers:{key:"grpc-trace-bin"  value:"AADYLLj/P24k5IPyxF69BSUsAYk6sKAoqXpiAgA"}  headers:{key:"x-forwarded-proto"  value:"https"}  headers:{key:"x-request-id"  value:"778b7a17-772d-42c9-96e8-e2b85d9dd325"}}} 
2022/10/21 17:00:40 Got RequestHeaders.Attributes map[]
2022/10/21 17:00:40 Got RequestHeaders.Headers headers:{key:":method"  value:"POST"}  headers:{key:":scheme"  value:"https"}  headers:{key:":path"  value:"/google.pubsub.v1.Publisher/Publish"}  headers:{key:":authority"  value:"pubsub.googleapis.com"}  headers:{key:"content-type"  value:"application/grpc"}  headers:{key:"user-agent"  value:"grpc-go/1.48.0"}  headers:{key:"te"  value:"trailers"}  headers:{key:"grpc-timeout"  value:"59870023u"}  headers:{key:"authorization"  value:"Bearer ya29.redacted"}  headers:{key:"grpc-tags-bin"  value:"AAAGc3RhdHVzAk9LAAV0b3BpYyhwcm9qZWN0cy9mYWJsZWQtcmF5LTEwNDExNy90b3BpY3MvdG9waWMx"}  headers:{key:"x-goog-api-client"  value:"gl-go/1.19.0 gccl/1.25.1 gapic/1.25.1 gax/2.4.0 grpc/1.48.0"}  headers:{key:"x-goog-request-params"  value:"topic=projects%2Ffabled-ray-104117%2Ftopics%2Ftopic1"}  headers:{key:"grpc-trace-bin"  value:"AADYLLj/P24k5IPyxF69BSUsAYk6sKAoqXpiAgA"}  headers:{key:"x-forwarded-proto"  value:"https"}  headers:{key:"x-request-id"  value:"778b7a17-772d-42c9-96e8-e2b85d9dd325"}
2022/10/21 17:00:40 Header :method POST
2022/10/21 17:00:40    RequestBody: J
(projects/fabled-ray-104117/topics/topic1

foo number 0

foo number 1
2022/10/21 17:00:40 >>>>>>>>>>>>>>>> Got message for topic: projects/fabled-ray-104117/topics/topic1
Decode PubsubMessage Data ---->  foo number 0
Decode PubsubMessage Data ---->  foo number 1
2022/10/21 17:00:40 pb.ProcessingRequest_ResponseHeaders &{headers:{headers:{key:":status"  value:"200"}  headers:{key:"content-disposition"  value:"attachment"}  headers:{key:"content-type"  value:"application/grpc"}  headers:{key:"date"  value:"Fri, 21 Oct 2022 21:00:40 GMT"}  headers:{key:"alt-svc"  value:"h3=\":443\"; ma=2592000,h3-29=\":443\"; ma=2592000,h3-Q050=\":443\"; ma=2592000,h3-Q046=\":443\"; ma=2592000,h3-Q043=\":443\"; ma=2592000,quic=\":443\"; ma=2592000; v=\"46,43\""}  headers:{key:"x-envoy-upstream-service-time"  value:"149"}}} 
2022/10/21 17:00:40 pb.ProcessingRequest_ResponseBody &{body:"\x00\x00\x00\x00$\n\x106016065317030900\n\x106016065317030901"  end_of_stream:true} 
2022/10/21 17:00:40    ResponseBody: $
6016065317030900
6016065317030901

```


### Envoy Logs

```log
[2022-10-21 17:00:40.611][2202332][debug][filter] [source/extensions/filters/listener/tls_inspector/tls_inspector.cc:116] tls:onServerName(), requestedServerName: pubsub.googleapis.com

[2022-10-21 17:00:40.611][2202332][debug][filter] [source/extensions/filters/listener/http_inspector/http_inspector.cc:53] http inspector: new connection accepted

[2022-10-21 17:00:40.611][2202332][debug][conn_handler] [source/server/active_tcp_listener.cc:142] [C2] new connection from 127.0.0.1:45076

[2022-10-21 17:00:40.614][2202336][debug][filter] [source/extensions/filters/listener/http_inspector/http_inspector.cc:53] http inspector: new connection accepted

[2022-10-21 17:00:40.752][2202333][debug][http] [source/common/http/conn_manager_impl.cc:299] [C6] new stream
[2022-10-21 17:00:40.752][2202333][debug][http] [source/common/http/conn_manager_impl.cc:904] [C6][S8431130007026407378] request headers complete (end_stream=false):
':method', 'POST'
':scheme', 'https'
':path', '/google.pubsub.v1.Publisher/Publish'
':authority', 'pubsub.googleapis.com'
'content-type', 'application/grpc'
'user-agent', 'grpc-go/1.48.0'
'te', 'trailers'
'grpc-timeout', '59870023u'
'authorization', 'Bearer ya29.redacted'
'grpc-tags-bin', 'AAAGc3RhdHVzAk9LAAV0b3BpYyhwcm9qZWN0cy9mYWJsZWQtcmF5LTEwNDExNy90b3BpY3MvdG9waWMx'
'x-goog-api-client', 'gl-go/1.19.0 gccl/1.25.1 gapic/1.25.1 gax/2.4.0 grpc/1.48.0'
'x-goog-request-params', 'topic=projects%2Ffabled-ray-104117%2Ftopics%2Ftopic1'
'grpc-trace-bin', 'AADYLLj/P24k5IPyxF69BSUsAYk6sKAoqXpiAgA'

[2022-10-21 17:00:40.753][2202333][debug][connection] [./source/common/network/connection_impl.h:89] [C6] current connecting state: false
[2022-10-21 17:00:40.753][2202333][debug][ext_proc] [source/extensions/filters/http/ext_proc/ext_proc.cc:90] Opening gRPC stream to external processor
[2022-10-21 17:00:40.753][2202333][debug][router] [source/common/router/router.cc:467] [C0][S2946864793608285876] cluster 'ext_proc_cluster' match for URL '/envoy.service.ext_proc.v3.ExternalProcessor/Process'
[2022-10-21 17:00:40.753][2202333][debug][router] [source/common/router/router.cc:670] [C0][S2946864793608285876] router decoding headers:
':method', 'POST'
':path', '/envoy.service.ext_proc.v3.ExternalProcessor/Process'
':authority', 'ext_proc_cluster'
':scheme', 'http'
'te', 'trailers'
'content-type', 'application/grpc'
'x-envoy-internal', 'true'
'x-forwarded-for', '192.168.1.178'

[2022-10-21 17:00:40.753][2202333][debug][pool] [source/common/http/conn_pool_base.cc:78] queueing stream due to no available connections (ready=0 busy=0 connecting=0)
[2022-10-21 17:00:40.753][2202333][debug][pool] [source/common/conn_pool/conn_pool_base.cc:268] trying to create new connection
[2022-10-21 17:00:40.753][2202333][debug][pool] [source/common/conn_pool/conn_pool_base.cc:145] creating a new connection (connecting=0)
[2022-10-21 17:00:40.753][2202333][debug][http2] [source/common/http/http2/codec_impl.cc:1783] [C10] updating connection-level initial window size to 268435456
[2022-10-21 17:00:40.753][2202333][debug][connection] [./source/common/network/connection_impl.h:89] [C10] current connecting state: true
[2022-10-21 17:00:40.753][2202333][debug][client] [source/common/http/codec_client.cc:57] [C10] connecting
[2022-10-21 17:00:40.753][2202333][debug][connection] [source/common/network/connection_impl.cc:912] [C10] connecting to 127.0.0.1:18080
[2022-10-21 17:00:40.754][2202333][debug][connection] [source/common/network/connection_impl.cc:931] [C10] connection in progress
[2022-10-21 17:00:40.754][2202333][debug][ext_proc] [source/extensions/filters/http/ext_proc/ext_proc.cc:141] Sending headers message
[2022-10-21 17:00:40.754][2202333][debug][http] [source/common/http/filter_manager.cc:841] [C6][S8431130007026407378] request end stream
[2022-10-21 17:00:40.754][2202333][debug][connection] [source/common/network/connection_impl.cc:683] [C10] connected
[2022-10-21 17:00:40.754][2202333][debug][client] [source/common/http/codec_client.cc:89] [C10] connected
[2022-10-21 17:00:40.754][2202333][debug][pool] [source/common/conn_pool/conn_pool_base.cc:305] [C10] attaching to next stream
[2022-10-21 17:00:40.754][2202333][debug][pool] [source/common/conn_pool/conn_pool_base.cc:177] [C10] creating stream
[2022-10-21 17:00:40.754][2202333][debug][router] [source/common/router/upstream_request.cc:422] [C0][S2946864793608285876] pool ready
[2022-10-21 17:00:40.756][2202333][debug][router] [source/common/router/router.cc:1351] [C0][S2946864793608285876] upstream headers complete: end_stream=false
[2022-10-21 17:00:40.756][2202333][debug][http] [source/common/http/async_client_impl.cc:101] async http request response headers (end_stream=false):
':status', '200'
'content-type', 'application/grpc'

[2022-10-21 17:00:40.757][2202333][debug][forward_proxy] [source/extensions/common/dynamic_forward_proxy/dns_cache_impl.cc:88] thread local lookup for host 'pubsub.googleapis.com'
[2022-10-21 17:00:40.825][2202325][debug][forward_proxy] [source/extensions/common/dynamic_forward_proxy/dns_cache_impl.cc:307] main thread resolve complete for host 'pubsub.googleapis.com': [142.251.163.95:0, 172.253.62.95:0, 172.253.122.95:0, 172.253.115.95:0, 142.250.31.95:0, 172.253.63.95:0, 142.251.16.95:0]
[2022-10-21 17:00:40.825][2202325][debug][forward_proxy] [source/extensions/common/dynamic_forward_proxy/dns_cache_impl.cc:374] host 'pubsub.googleapis.com' address has changed from <empty> to 142.251.163.95:443

[2022-10-21 17:00:40.825][2202325][debug][upstream] [source/extensions/clusters/dynamic_forward_proxy/cluster.cc:112] Adding host info for pubsub.googleapis.com
[2022-10-21 17:00:40.826][2202333][debug][router] [source/common/router/router.cc:467] [C6][S8431130007026407378] cluster 'dynamic_forward_proxy_cluster' match for URL '/google.pubsub.v1.Publisher/Publish'
[2022-10-21 17:00:40.826][2202333][debug][router] [source/common/router/router.cc:670] [C6][S8431130007026407378] router decoding headers:
':method', 'POST'
':scheme', 'https'
':path', '/google.pubsub.v1.Publisher/Publish'
':authority', 'pubsub.googleapis.com'
'content-type', 'application/grpc'
'user-agent', 'grpc-go/1.48.0'
'te', 'trailers'
'grpc-timeout', '59870023u'
'authorization', 'Bearer ya29.redacted'
'grpc-tags-bin', 'AAAGc3RhdHVzAk9LAAV0b3BpYyhwcm9qZWN0cy9mYWJsZWQtcmF5LTEwNDExNy90b3BpY3MvdG9waWMx'
'x-goog-api-client', 'gl-go/1.19.0 gccl/1.25.1 gapic/1.25.1 gax/2.4.0 grpc/1.48.0'
'x-goog-request-params', 'topic=projects%2Ffabled-ray-104117%2Ftopics%2Ftopic1'
'grpc-trace-bin', 'AADYLLj/P24k5IPyxF69BSUsAYk6sKAoqXpiAgA'
'x-forwarded-proto', 'https'
'x-request-id', '778b7a17-772d-42c9-96e8-e2b85d9dd325'
'x-envoy-expected-rq-timeout-ms', '15000'

[2022-10-21 17:00:40.979][2202333][debug][http] [source/common/http/conn_manager_impl.cc:1516] [C6][S8431130007026407378] encoding headers via codec (end_stream=false):
':status', '200'
'content-disposition', 'attachment'
'content-type', 'application/grpc'
'date', 'Fri, 21 Oct 2022 21:00:40 GMT'
'alt-svc', 'h3=":443"; ma=2592000,h3-29=":443"; ma=2592000,h3-Q050=":443"; ma=2592000,h3-Q046=":443"; ma=2592000,h3-Q043=":443"; ma=2592000,quic=":443"; ma=2592000; v="46,43"'
'x-envoy-upstream-service-time', '149'
'server', 'envoy'

[2022-10-21 17:00:40.979][2202333][debug][http] [source/common/http/conn_manager_impl.cc:1533] [C6][S8431130007026407378] encoding trailers via codec:
'grpc-status', '0'
'content-disposition', 'attachment'

[2022-10-21 17:00:40.979][2202333][debug][http2] [source/common/http/http2/codec_impl.cc:1470] [C6] stream 1 closed: 0

```

