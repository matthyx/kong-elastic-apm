# Kong Elastic APM Plugin


[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/6092/badge)](https://bestpractices.coreinfrastructure.org/projects/6092) ![Stars](https://img.shields.io/github/stars/matthyx/kong-elastic-apm?style=flat-square) ![Version](https://img.shields.io/github/v/release/Kong/kong?color=green&label=Version&style=flat-square)  ![License](https://img.shields.io/badge/License-Apache%202.0-blue?style=flat-square)

This plugin enable the open telemetry feature on the Kong gateway to feed the Elastic Application Performance Monitoring solution. This plugin written in Go, leverage the [Elastic APM Go agent](https://www.elastic.co/guide/en/apm/agent/go/current/index.html).

To set up this plugin please follow the instructions of the APM Go agent, and refers to the environment variables setup, as there are mapped from the plugin configuration (to lower case).


**Kong** or **Kong API Gateway** is a cloud-native, platform-agnostic, scalable API Gateway distinguished for its high performance and extensibility via plugins.

By providing functionality for proxying, routing, load balancing, health checking, authentication, Kong serves as the central layer for orchestrating microservices or conventional API traffic with ease.

Kong runs natively on Kubernetes thanks to its official [Kubernetes Ingress Controller](https://github.com/Kong/kubernetes-ingress-controller).

**Elastic Observability** Unified visibility across your entire ecosystem
Bring your logs, metrics, and APM traces together at scale in a single stack, so you can monitor and react to events happening anywhere in your environment. And it's free and open. [Elastic](https://www.elastic.co/observability)

## License

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Trying the plugin

Build the image:
```shell
docker-compose build
```

Run the image:
```shell
docker-compose up
```

Test the service:
```shell
curl -v http://localhost:8000/hello
```

You can now check APM metrics in Kibana: http://localhost:5601/app/apm
