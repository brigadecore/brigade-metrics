load('ext://min_k8s_version', 'min_k8s_version')
min_k8s_version('1.18.0')

trigger_mode(TRIGGER_MODE_MANUAL)

load('ext://namespace', 'namespace_create')
namespace_create('brigade-metrics')
k8s_resource(
  new_name = 'namespace',
  objects = ['brigade-metrics:namespace'],
  labels = ['brigade-metrics']
)

docker_build(
  'brigadecore/brigade-metrics-exporter', '.',
  dockerfile = 'exporter/Dockerfile',
  only = [
    'exporter/',
    'go.mod',
    'go.sum'
  ],
  ignore = ['**/*_test.go']
)
k8s_resource(
  workload = 'brigade-metrics-exporter',
  new_name = 'exporter',
  labels = ['brigade-metrics']
)
k8s_resource(
  workload = 'exporter',
  objects = ['brigade-metrics-exporter:secret']
)

docker_build(
  'brigadecore/brigade-metrics-grafana', '.',
  dockerfile = 'grafana/Dockerfile',
  only = ['grafana/']
)
k8s_resource(
  workload = 'brigade-metrics-grafana',
  new_name = 'grafana',
  port_forwards = '31700:80',
  labels = ['brigade-metrics']
)
k8s_resource(
  workload = 'grafana',
  objects = [
    'brigade-metrics-grafana:persistentvolumeclaim',
    'brigade-metrics-grafana:secret',
    'brigade-metrics-grafana-datasources:configmap',
    'brigade-metrics-nginx:secret'
  ]
)

k8s_resource(
  workload = 'brigade-metrics-prometheus',
  new_name = 'prometheus',
  labels = ['brigade-metrics'],
)
k8s_resource(
  workload = 'prometheus',
  objects = [
    'brigade-metrics-prometheus:configmap',
    'brigade-metrics-prometheus:persistentvolumeclaim'
  ]
)

k8s_yaml(
  helm(
    './charts/brigade-metrics',
    name = 'brigade-metrics',
    namespace = 'brigade-metrics',
    set = [
      'receiver.tls.enabled=false',
      'exporter.brigade.apiToken=' + os.environ['BRIGADE_API_TOKEN'],
      'grafana.auth.username=admin',
      'grafana.auth.password=admin',
      'grafana.tls.enabled=false'
    ]
  )
)

