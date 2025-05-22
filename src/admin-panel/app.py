from flask import Flask, render_template_string, request, jsonify
import requests

app = Flask(__name__)

SERVICES = {
    "postgresql": {"port": 5434, "auth": False},
    "kafka": {"port": 9092, "auth": False},
    "redis": {"port": 6379, "auth": False}
}

HTML_TEMPLATE = """
<!doctype html>
<html>
<head><title>Auth Admin</title></head>
<body>
    <h1>Services Control</h1>
    <table border=1>
        <tr><th>Service</th><th>Auth Status</th><th>Action</th></tr>
        {% for name, data in services.items() %}
        <tr>
            <td>{{ name }}</td>
            <td id="status-{{ name }}">{{ "Enabled" if data.auth else "Disabled" }}</td>
            <td>
                <button onclick="toggleAuth('{{ name }}')">Toggle</button>
            </td>
        </tr>
        {% endfor %}
    </table>
    <script>
    function toggleAuth(service) {
        fetch('/toggle/' + service)
        .then(r => r.json())
        .then(data => {
            document.getElementById('status-' + service).innerText =
                data.auth_enabled ? "Enabled" : "Disabled";
        });
    }
    </script>
</body>
</html>
"""

@app.route('/')
def index():
    # Обновляем статусы
    for name in SERVICES:
        try:
            resp = requests.get(f"http://{name}:8080/admin/config",
                              headers={"Authorization": "Bearer admin-secret-token"})
            SERVICES[name]["auth"] = resp.json().get("auth_enabled", False)
        except:
            SERVICES[name]["auth"] = False
    return render_template_string(HTML_TEMPLATE, services=SERVICES)

@app.route('/toggle/<service>')
def toggle(service):
    if service not in SERVICES:
        return jsonify({"error": "Service not found"}), 404

    try:
        resp = requests.post(f"http://{service}:8080/admin/auth/toggle",
                           headers={"Authorization": "Bearer admin-secret-token"})
        return resp.json()
    except Exception as e:
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
