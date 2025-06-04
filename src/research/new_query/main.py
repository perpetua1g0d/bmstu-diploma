import csv
import matplotlib.pyplot as plt
import numpy as np
from collections import defaultdict
import matplotlib as mpl
from matplotlib.ticker import FormatStrFormatter

# Настройка стиля графиков
plt.style.use('seaborn-v0_8-whitegrid')
mpl.rcParams['font.family'] = 'DejaVu Sans'
mpl.rcParams['axes.labelsize'] = 12
mpl.rcParams['xtick.labelsize'] = 10
mpl.rcParams['ytick.labelsize'] = 10
mpl.rcParams['legend.fontsize'] = 10
mpl.rcParams['figure.titlesize'] = 14
mpl.rcParams['hatch.linewidth'] = 0.7

def parse_duration(duration_str):
    """Преобразует строку с временем в миллисекунды"""
    if not duration_str:
        return 0.0

    if duration_str.endswith('ns'):
        return float(duration_str[:-2]) / 1_000_000
    elif duration_str.endswith('µs'):
        return float(duration_str[:-2]) / 1000
    elif duration_str.endswith('ms'):
        return float(duration_str[:-2])
    elif duration_str.endswith('s'):
        if 'm' in duration_str:  # Для формата типа "1m23.456s"
            parts = duration_str[:-1].split('m')
            if len(parts) == 2:
                return float(parts[0]) * 60000 + float(parts[1]) * 1000
        return float(duration_str[:-1]) * 1000
    else:
        try:
            return float(duration_str)
        except ValueError:
            return 0.0

# Чтение и обработка данных с усреднением
def process_file(filename):
    """Обрабатывает файл результатов и возвращает усредненные данные"""
    data_dict = defaultdict(lambda: defaultdict(list))

    try:
        with open(filename, 'r') as file:
            reader = csv.reader(file)
            for row in reader:
                if len(row) < 10:
                    continue

                query_type = row[3]
                total_requests = int(row[1])
                avg_duration = parse_duration(row[7])

                # Добавляем значение в словарь для усреднения
                data_dict[query_type][total_requests].append(avg_duration)
    except FileNotFoundError:
        print(f"Файл не найден: {filename}")
    except Exception as e:
        print(f"Ошибка при обработке файла {filename}: {str(e)}")

    return data_dict

# Обрабатываем оба файла
disabled_data = process_file('results_disabled.csv')
auth_data = process_file('results_auth.csv')

# Подготовка усредненных данных для сравнения
comparison_data = defaultdict(list)

# Собираем данные для light запросов
if 'light' in disabled_data:
    for req, durations in disabled_data['light'].items():
        avg = sum(durations) / len(durations)
        comparison_data['light'].append((req, 'Без аутентификации', avg))

if 'light' in auth_data:
    for req, durations in auth_data['light'].items():
        avg = sum(durations) / len(durations)
        comparison_data['light'].append((req, 'С аутентификацией', avg))

# Собираем данные для heavy запросов
if 'heavy' in disabled_data:
    for req, durations in disabled_data['heavy'].items():
        avg = sum(durations) / len(durations)
        comparison_data['heavy'].append((req, 'Без аутентификации', avg))

if 'heavy' in auth_data:
    for req, durations in auth_data['heavy'].items():
        avg = sum(durations) / len(durations)
        comparison_data['heavy'].append((req, 'С аутентификацией', avg))

# Функция для создания гистограмм сравнения
def plot_comparison(data, title):
    if not data:
        print(f"⚠️ Нет данных для {title}")
        return

    # Группируем данные по количеству запросов
    request_groups = sorted(set(req for req, _, _ in data))
    disabled_values = []
    auth_values = []

    for req in request_groups:
        # Получаем значения для текущего количества запросов
        disabled_vals = [val for (r, mode, val) in data if r == req and mode == 'Без аутентификации']
        auth_vals = [val for (r, mode, val) in data if r == req and mode == 'С аутентификацией']

        disabled_values.append(disabled_vals[0] if disabled_vals else 0)
        auth_values.append(auth_vals[0] if auth_vals else 0)

    x = np.arange(len(request_groups))
    width = 0.35

    fig, ax = plt.subplots(figsize=(12, 7))
    fig.suptitle(f'Сравнение времени выполнения запросов: {title}', fontsize=14, fontweight='bold')

    # Создаем столбцы с улучшенным оформлением
    rects1 = ax.bar(
        x - width/2,
        disabled_values,
        width,
        label='Без аутентификации',
        color='#4caf50',  # Зеленый
        edgecolor='darkgreen',
        linewidth=1.2,
        alpha=0.9
    )

    rects2 = ax.bar(
        x + width/2,
        auth_values,
        width,
        label='С аутентификацией',
        color='#ff7043',  # Оранжево-красный
        hatch='////',
        edgecolor='#c43e1c',
        linewidth=1.2,
        alpha=0.9
    )

    # Настройка оформления
    ax.set_xlabel('Общее количество запросов', fontsize=12, labelpad=10)
    ax.set_ylabel('Среднее время выполнения (мс)', fontsize=12, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels([f"{req:,}" for req in request_groups])

    # Форматирование значений на оси Y
    ax.yaxis.set_major_formatter(FormatStrFormatter('%.2f'))

    # Добавляем значения над столбцами
    def autolabel(rects):
        for rect in rects:
            height = rect.get_height()
            ax.annotate(f'{height:.2f}',
                        xy=(rect.get_x() + rect.get_width() / 2, height),
                        xytext=(0, 3),
                        textcoords="offset points",
                        ha='center', va='bottom',
                        fontsize=9,
                        bbox=dict(boxstyle="round,pad=0.1",
                                  fc="white",
                                  ec="gray",
                                  alpha=0.7))

    autolabel(rects1)
    autolabel(rects2)

    # Настраиваем легенду
    ax.legend(frameon=True, framealpha=0.9, loc='best')

    # Добавляем сетку
    ax.grid(True, linestyle='--', alpha=0.6, axis='y')

    # Добавляем горизонтальную линию для нуля
    ax.axhline(y=0, color='gray', linewidth=0.8)

    # Рассчитываем и отображаем разницу в производительности
    for i, req in enumerate(request_groups):
        if disabled_values[i] > 0 and auth_values[i] > 0:
            diff = ((auth_values[i] / disabled_values[i]) - 1) * 100
            ax.text(i, max(disabled_values[i], auth_values[i]) * 1.05,
                    f'+{diff:.1f}%' if diff > 0 else f'{diff:.1f}%',
                    ha='center', va='bottom',
                    fontsize=9,
                    bbox=dict(boxstyle="round,pad=0.2",
                              fc="yellow",
                              ec="gold",
                              alpha=0.7))

    # Автоматическая настройка макета
    plt.tight_layout(rect=[0, 0, 1, 0.96])

    # Сохраняем график в файл
    plt.savefig(f'{title}_comparison.png', dpi=300, bbox_inches='tight')
    plt.show()

# Создаем гистограммы сравнения
if comparison_data:
    if 'light' in comparison_data:
        plot_comparison(comparison_data['light'], 'Лёгкие запросы')
    if 'heavy' in comparison_data:
        plot_comparison(comparison_data['heavy'], 'Тяжёлые запросы')
else:
    print("Нет данных .")
