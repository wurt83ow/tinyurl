CREATE TABLE IF NOT EXISTS dataurl (
	correlation_id VARCHAR(50) PRIMARY KEY, 
	short_url TEXT,
	user_id VARCHAR(50) REFERENCES users (Id) NOT NULL,
	original_url TEXT 	 
	);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_url ON dataurl (original_url);