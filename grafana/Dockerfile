FROM grafana/grafana:8.0.2

ENV GF_PATHS_DATA=/var/lib/grafana/data
ENV GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH=/var/lib/grafana/dashboards/brigade.json
ENV GF_USERS_ALLOW_SIGN_UP=false

COPY grafana/dashboards/ /var/lib/grafana/dashboards/
