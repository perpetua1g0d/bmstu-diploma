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
    sign = "true" if request.form.get('sign') else "false"
    verify = "true" if request.form.get('verify') else "false"

    try:
        # Обновляем ConfigMap в неймспейсе сервиса
        cm = v1.read_namespaced_config_map("auth-settings", service)
        cm.data["SIGN_AUTH_ENABLED"] = sign
        cm.data["VERIFY_AUTH_ENABLED"] = verify
        v1.replace_namespaced_config_map("auth-settings", service, cm)

        # Обновляем кэш
        settings_cache[service] = {
            "sign": sign == "true",
            "verify": verify == "true"
        }

        # Возвращаемся на главную с сообщением
        return redirect(url_for(
            'index',
            service=service,
            message=f"Настройки для {service} обновлены!"
        ))
    except Exception as e:
        return redirect(url_for(
            'index',
            service=service,
            message=f"Ошибка: {str(e)}"
        ))

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
