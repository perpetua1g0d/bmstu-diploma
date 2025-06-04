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
    if duration_str.endswith('µs'):
        return float(duration_str[:-2]) / 1000
    elif duration_str.endswith('ms'):
        return float(duration_str[:-2])
    elif duration_str.endswith('s'):
        if 'm' in duration_str:  # Для формата типа "1.234s"
            parts = duration_str[:-1].split('m')
            if len(parts) == 2:
                return float(parts[0]) * 60000 + float(parts[1]) * 1000
        return float(duration_str[:-1]) * 1000
    else:
        return float(duration_str)

# Чтение и обработка данных с усреднением
data_dict = defaultdict(lambda: defaultdict(list))

with open('results_disabled.csv', 'r') as file:
    reader = csv.reader(file)
    for row in reader:
        if len(row) < 10:
            continue

        query_type = row[3]
        total_requests = int(row[1])
        avg_duration = parse_duration(row[7])

        # Добавляем значение в словарь для усреднения
        data_dict[query_type][total_requests].append(avg_duration)

# Подготовка усредненных данных
light_data = []
heavy_data = []

for query_type, requests_data in data_dict.items():
    for total_requests, durations in requests_data.items():
        avg_duration = sum(durations) / len(durations)
        if query_type == 'light':
            light_data.append((total_requests, avg_duration))
        elif query_type == 'heavy':
            heavy_data.append((total_requests, avg_duration))

# Сортировка данных по количеству запросов
light_data.sort(key=lambda x: x[0])
heavy_data.sort(key=lambda x: x[0])

# Функция для создания улучшенных гистограмм
def plot_benchmark(data, title):
    if not data:
        print(f"No data available for {title}")
        return

    total_requests = sorted(set([d[0] for d in data]))
    real_avgs = []
    fake_avgs = []

    for req in total_requests:
        # Находим среднее значение для данного количества запросов
        avg = next(d[1] for d in data if d[0] == req)
        real_avgs.append(avg)
        fake_avgs.append(avg * 0.9)  # "Фейковые" данные - 90% от реальных

    x = np.arange(len(total_requests))
    width = 0.35

    fig, ax = plt.subplots(figsize=(12, 7))
    fig.suptitle(f'Среднее время выполнения запросов ({title})', fontsize=14, fontweight='bold')

    # Создаем столбцы с улучшенным оформлением
    rects1 = ax.bar(
        x - width/2,
        real_avgs,
        width,
        label='Реальные данные',
        color='#4caf50',  # Приятный зеленый
        edgecolor='darkgreen',
        linewidth=1.2,
        alpha=0.9
    )

    rects2 = ax.bar(
        x + width/2,
        fake_avgs,
        width,
        label='Теоретические данные (90% от реальных)',
        color='#ff7043',  # Приятный оранжево-красный
        hatch='////',
        edgecolor='#c43e1c',
        linewidth=1.2,
        alpha=0.9
    )

    # Настройка оформления
    ax.set_xlabel('Общее количество запросов', fontsize=12, labelpad=10)
    ax.set_ylabel('Среднее время выполнения (мс)', fontsize=12, labelpad=10)
    ax.set_xticks(x)
    ax.set_xticklabels([f"{req:,}" for req in total_requests])

    # Форматирование значений на оси Y
    ax.yaxis.set_major_formatter(FormatStrFormatter('%.2f'))

    # Добавляем значения над столбцами
    def autolabel(rects):
        for rect in rects:
            height = rect.get_height()
            ax.annotate(f'{height:.2f}',
                        xy=(rect.get_x() + rect.get_width() / 2, height),
                        xytext=(0, 3),  # 3 points vertical offset
                        textcoords="offset points",
                        ha='center', va='bottom',
                        fontsize=9)

    autolabel(rects1)
    autolabel(rects2)

    # Настраиваем легенду
    ax.legend(frameon=True, framealpha=0.9, loc='upper left')

    # Добавляем сетку
    ax.grid(True, linestyle='--', alpha=0.7)

    # Автоматическая настройка макета с дополнительным пространством сверху
    plt.tight_layout(rect=[0, 0, 1, 0.96])

    # Сохраняем график в файл
    plt.savefig(f'{title}_benchmark.png', dpi=300)
    plt.show()

# Создаем гистограммы
plot_benchmark(light_data, 'Лёгкие запросы')
plot_benchmark(heavy_data, 'Тяжёлые запросы')
