-- wrk Lua script for load testing URL shortener
-- This script simulates realistic mixed workload

-- Initialize random seed
math.randomseed(os.time())

-- Sample URLs for shortening
local urls = {
    "https://example.com/page1",
    "https://example.com/page2",
    "https://example.com/page3",
    "https://github.com/user/repo",
    "https://stackoverflow.com/questions/12345",
    "https://news.ycombinator.com/item?id=12345",
    "https://reddit.com/r/programming",
    "https://twitter.com/user/status/123456",
    "https://medium.com/@user/article-title",
    "https://youtube.com/watch?v=abcdefg"
}

-- Sample short codes (for redirect testing)
local shortCodes = {
    "abc123",
    "xyz789",
    "test01",
    "demo01",
    "link01"
}

-- Request counter
local counter = 0

-- Request method
request = function()
    counter = counter + 1
    local rand = math.random(1, 100)

    -- 70% GET requests (redirects)
    if rand <= 70 then
        local shortCode = shortCodes[math.random(1, #shortCodes)]
        return wrk.format("GET", "/" .. shortCode)

    -- 20% POST requests (create short URL)
    elseif rand <= 90 then
        local url = urls[math.random(1, #urls)]
        local body = string.format('{"url":"%s"}', url)
        local headers = {
            ["Content-Type"] = "application/json"
        }
        return wrk.format("POST", "/api/v1/shorten", headers, body)

    -- 10% GET analytics
    else
        local shortCode = shortCodes[math.random(1, #shortCodes)]
        return wrk.format("GET", "/api/v1/analytics/" .. shortCode .. "?days=30")
    end
end

-- Response callback
response = function(status, headers, body)
    if status >= 400 then
        print("Error response: " .. status)
    end
end

-- Done callback (print summary)
done = function(summary, latency, requests)
    io.write("------------------------------\n")
    io.write("Load Test Summary\n")
    io.write("------------------------------\n")
    io.write(string.format("Total Requests: %d\n", summary.requests))
    io.write(string.format("Total Duration: %.2fs\n", summary.duration / 1000000))
    io.write(string.format("Requests/sec: %.2f\n", summary.requests / (summary.duration / 1000000)))
    io.write(string.format("Avg Latency: %.2fms\n", latency.mean / 1000))
    io.write(string.format("Max Latency: %.2fms\n", latency.max / 1000))
    io.write(string.format("Stdev Latency: %.2fms\n", latency.stdev / 1000))
    io.write(string.format("P50 Latency: %.2fms\n", latency:percentile(50) / 1000))
    io.write(string.format("P90 Latency: %.2fms\n", latency:percentile(90) / 1000))
    io.write(string.format("P95 Latency: %.2fms\n", latency:percentile(95) / 1000))
    io.write(string.format("P99 Latency: %.2fms\n", latency:percentile(99) / 1000))
    io.write(string.format("Errors: %d\n", summary.errors.connect + summary.errors.read + summary.errors.write + summary.errors.timeout))
    io.write("------------------------------\n")
end
