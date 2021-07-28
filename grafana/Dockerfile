FROM grafana/grafana:8.0.2

ENV GF_PATHS_DATA=/var/lib/grafana/data
ENV GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH=/var/lib/grafana/dashboards/brigade.json
ENV GF_USERS_ALLOW_SIGN_UP=false
ENV GF_DASHBOARDS_MIN_REFRESH_INTERVAL=2s


# TODO: Not sure if the license permits this. Run this by CELA.
COPY grafana/images/ /usr/share/grafana/public/img/

COPY grafana/dashboards/ /var/lib/grafana/dashboards/
