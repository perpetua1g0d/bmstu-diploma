from flask import Flask, request, render_template
from kubernetes import client, config
import os

app = Flask(__name__)
config.load_incluster_config()
v1 = client.CoreV1Api()

# Список доступных сервисов из переменных окружения
SERVICES = os.getenv("SERVICES", "postgres-a,postgres-b").split(',')

@app.route('/')
def index():
    return render_template('index.html', services=SERVICES)

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
        return f"Настройки для {service} обновлены!"
    except Exception as e:
        return f"Ошибка: {str(e)}", 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
