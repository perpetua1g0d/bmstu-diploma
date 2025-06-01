from flask import Flask, request, render_template, jsonify, redirect, url_for
from kubernetes import client, config
import os
import requests

app = Flask(__name__)
app.secret_key = os.getenv("FLASK_SECRET", "supersecretkey")
config.load_incluster_config()
v1 = client.CoreV1Api()

# Список доступных сервисов из переменных окружения
SERVICES = os.getenv("SERVICES", "postgres-a,postgres-b").split(',')

# Кэш для хранения текущих настроек
settings_cache = {service: {"sign": True, "verify": True} for service in SERVICES}

IDP_SERVICE_URL = "http://idp.idp.svc.cluster.local:80"

def fetch_current_settings(service):
    try:
        cm = v1.read_namespaced_config_map("auth-settings", service)
        sign_enabled = cm.data.get("SIGN_AUTH_ENABLED", "true") == "true"
        verify_enabled = cm.data.get("VERIFY_AUTH_ENABLED", "true") == "true"
        return sign_enabled, verify_enabled
    except Exception as e:
        app.logger.error(f"Error fetching settings for {service}: {str(e)}")
        return True, True  # Значения по умолчанию

@app.route('/')
def index():
    try:
        selected_service = request.args.get('service', SERVICES[0])
        app.logger.info(f"Fetching settings for: {selected_service}")
        sign_enabled, verify_enabled = fetch_current_settings(selected_service)
        app.logger.info(f"Settings: sign={sign_enabled}, verify={verify_enabled}")

        return render_template(
            'index.html',
            services=SERVICES,
            selected_service=selected_service,
            sign_enabled=sign_enabled,
            verify_enabled=verify_enabled,
            message=request.args.get('message', '')
        )
    except Exception as e:
        app.logger.exception("Error in index route")
        return f"Internal Server Error: {str(e)}", 500

@app.route('/settings')
def get_settings():
    service = request.args.get('service', SERVICES[0])
    sign_enabled, verify_enabled = fetch_current_settings(service)
    return jsonify({
        "sign": sign_enabled,
        "verify": verify_enabled
    })

@app.route('/update', methods=['POST'])
def update():
    service = request.form['service']
    sign = request.form.get('sign') == 'on'
    verify = request.form.get('verify') == 'on'

    try:
        # Обновляем ConfigMap
        cm = v1.read_namespaced_config_map("auth-settings", service)
        cm.data["SIGN_AUTH_ENABLED"] = str(sign).lower()
        cm.data["VERIFY_AUTH_ENABLED"] = str(verify).lower()
        v1.replace_namespaced_config_map("auth-settings", service, cm)

        # Отправляем уведомление сайдкарам
        notify_sidecars(service, sign, verify)

        return redirect(url_for('index', service=service, message=f"Настройки авторизации для {service} успешно применены."))
    except Exception as e:
        return redirect(url_for('index', service=service, message=f"Ошибка: {str(e)}"))

@app.route('/update_all', methods=['POST'])
def update_all():
    sign = request.form.get('sign_all') == 'on'
    verify = request.form.get('verify_all') == 'on'

    try:
        for service in SERVICES:
            # Обновляем ConfigMap для каждого сервиса
            cm = v1.read_namespaced_config_map("auth-settings", service)
            cm.data["SIGN_AUTH_ENABLED"] = str(sign).lower()
            cm.data["VERIFY_AUTH_ENABLED"] = str(verify).lower()
            v1.replace_namespaced_config_map("auth-settings", service, cm)

            # Отправляем уведомление сайдкарам
            notify_sidecars(service, sign, verify)

        return redirect(url_for('index', message="Глобальные настройки авторизации для всех сервисов успешно применены."))
    except Exception as e:
        return redirect(url_for('index', message=f"Ошибка: {str(e)}"))

@app.route('/get_permissions', methods=['POST'])
def get_permissions():
    client_id = request.form.get('client')
    scope = request.form.get('scope')

    if not client_id or not scope:
        return jsonify({"error": "Клиент или scope не должны быть пустыми"}), 400

    try:
        response = requests.get(
            f"{IDP_SERVICE_URL}/get_permissions",
            json={"client": client_id, "scope": scope},
            timeout=5
        )
        return jsonify(response.json()), response.status_code
    except Exception as e:
        return jsonify({"error": str(e)}), 500

@app.route('/update_permissions', methods=['POST'])
def update_permissions():
    client_id = request.form.get('client')
    scope = request.form.get('scope')
    roles = request.form.get('roles', '').split(',')

    if not client_id or not scope or not roles:
        return jsonify({"error": "Клиент, scope или роли пустые"}), 400

    try:
        response = requests.post(
            f"{IDP_SERVICE_URL}/update_permissions",
            json={"client": client_id, "scope": scope, "roles": roles},
            timeout=5
        )
        return jsonify({"message": "Права успешно применены"}), response.status_code
    except Exception as e:
        return jsonify({"error": str(e)}), 500

def notify_sidecars(service, sign, verify):
    pods = v1.list_namespaced_pod(namespace=service, label_selector=f"app={service}")

    for pod in pods.items:
        try:
            data = {"sign": sign, "verify": verify}
            requests.post(
                f"http://{pod.status.pod_ip}:8080/reload-config",
                json=data,
                timeout=1
            )
            app.logger.info(f"Notification sent to {pod.metadata.name}: {data}")
        except Exception as e:
            app.logger.error(f"Failed to notify {pod.metadata.name}: {str(e)}")

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
