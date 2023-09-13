CREATE TABLE IF NOT EXISTS dataurl (
	correlation_id VARCHAR(50) PRIMARY KEY, 
	short_url TEXT,
	original_url TEXT 	 
	);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_url ON dataurl (original_url);