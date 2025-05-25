import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns

sns.set(style="whitegrid", rc={"grid.linestyle": ":"})
plt.figure(figsize=(12, 7))

df_disabled = pd.read_csv("auth_disabled.csv")
df_enabled = pd.read_csv("auth_enabled.csv")

df = pd.concat([
    df_disabled.assign(auth="Выключена"),
    df_enabled.assign(auth="Включена")
])

lineplot = sns.lineplot(
    data=df,
    x="requests",
    y="time_ms",
    hue="auth",
    style="auth",
    markers={"Выключена": "o", "Включена": "s"},
    dashes={"Выключена": (4, 2), "Включена": ""},
    palette="husl",
    markersize=10,
    linewidth=2.5,
    legend="brief"
)

plt.title("Влияние авторизации на время выполнения запросов на запись", pad=20, fontsize=14)
plt.xlabel("Количество запросов", fontsize=12)
plt.ylabel("Среднее время, мс", fontsize=12)
plt.xticks(df["requests"].unique())

handles, labels = lineplot.get_legend_handles_labels()
plt.legend(
    handles=handles[2:],
    labels=labels[2:],
    title="Авторизация",
    title_fontsize=12,
    loc="upper left",
    frameon=True,
    shadow=True
)

for index, row in df.iterrows():
    plt.annotate(
        f"{row['time_ms']:.2f}",
        (row['requests'], row['time_ms']),
        textcoords="offset points",
        xytext=(0,8),
        ha="center",
        fontsize=9
    )

plt.tight_layout()
plt.savefig("lineplot_auth.png", dpi=150)
plt.show()
