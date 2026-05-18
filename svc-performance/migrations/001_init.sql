-- Initialize Performance schema
CREATE SCHEMA IF NOT EXISTS performance;

-- Performance events table
CREATE TABLE IF NOT EXISTS performance.events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES video.videos(id),
    event_type VARCHAR(32) NOT NULL,
    source VARCHAR(32),
    metadata JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Daily analytics table
CREATE TABLE IF NOT EXISTS performance.analytics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    video_id UUID NOT NULL REFERENCES video.videos(id),
    date DATE NOT NULL,
    views BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    watch_time BIGINT DEFAULT 0,
    avg_view_duration DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(video_id, date)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_performance_events_video ON performance.events(video_id);
CREATE INDEX IF NOT EXISTS idx_performance_events_timestamp ON performance.events(timestamp);
CREATE INDEX IF NOT EXISTS idx_performance_events_type ON performance.events(event_type);
CREATE INDEX IF NOT EXISTS idx_performance_analytics_video ON performance.analytics(video_id);
CREATE INDEX IF NOT EXISTS idx_performance_analytics_date ON performance.analytics(date);

-- Partitioned events table for better performance (optional)
-- CREATE TABLE IF NOT EXISTS performance.events_y2024m01 PARTITION OF performance.events FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Grant permissions
GRANT USAGE ON SCHEMA performance TO videoforge;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA performance TO videoforge;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA performance TO videoforge;