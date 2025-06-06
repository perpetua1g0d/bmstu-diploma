WITH heavy_cte AS (SELECT generate_series(1,1000000) AS data)
	SELECT COUNT(*), AVG(data) FROM heavy_cte
