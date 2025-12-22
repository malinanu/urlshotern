# Weekly Performance Review

**Date:** [YYYY-MM-DD]
**Week Number:** [WW/YYYY]
**Facilitator:** [Name]
**Attendees:** [List]

---

## 1. Key Metrics Summary

### Week-over-Week Comparison

| Metric | Previous Week | This Week | Change | Status |
|--------|--------------|-----------|--------|--------|
| URL Redirect Latency (p95) | XXms | YYms | +/-Z% | [GREEN/YELLOW/RED] |
| URL Shortening Latency (p95) | XXms | YYms | +/-Z% | [GREEN/YELLOW/RED] |
| Cache Hit Rate | XX% | YY% | +/-Z% | [GREEN/YELLOW/RED] |
| Database Query Time (p95) | XXms | YYms | +/-Z% | [GREEN/YELLOW/RED] |
| API Throughput | XXk req/s | YYk req/s | +/-Z% | [GREEN/YELLOW/RED] |
| Error Rate | X.XX% | Y.YY% | +/-Z% | [GREEN/YELLOW/RED] |
| Infrastructure Cost | $XXX | $YYY | +/-Z% | [GREEN/YELLOW/RED] |

**Overall Status:** [GREEN / YELLOW / RED]

### Performance Budget Compliance

| Endpoint | Budget | Actual (p95) | Status | Notes |
|----------|--------|--------------|--------|-------|
| GET /{shortCode} | < 50ms | XXms | [PASS/FAIL] | |
| POST /api/v1/shorten | < 200ms | XXms | [PASS/FAIL] | |
| GET /api/v1/analytics | < 500ms | XXms | [PASS/FAIL] | |

---

## 2. Alerts and Incidents

### Performance Alerts This Week

| Date/Time | Alert Type | Severity | Duration | Resolution | Owner |
|-----------|------------|----------|----------|------------|-------|
| | | | | | |

### Incidents Requiring Post-Mortem

| Incident ID | Description | Impact | Status | Post-Mortem Due |
|-------------|-------------|--------|--------|----------------|
| | | | | |

### Notable Anomalies

**Anomaly 1:**
- Description: [What was observed]
- Impact: [User impact if any]
- Investigation Status: [Resolved / In Progress / To Be Investigated]
- Root Cause: [If known]
- Action Items: [If any]

---

## 3. In-Flight Optimizations

### Active Work

| Optimization ID | Title | Owner | Status | Progress | ETA |
|-----------------|-------|-------|--------|----------|-----|
| | | | | XX% | |

### A/B Test Results

**Test ID:** [ID]
**Test Name:** [Name]
**Status:** [Running / Completed]
**Duration:** [X days]
**Sample Size:** [N sessions]

| Variant | Traffic | Latency (p95) | Error Rate | Statistical Significance |
|---------|---------|---------------|------------|-------------------------|
| Control | XX% | XXms | X.XX% | - |
| Variant A | XX% | XXms | X.XX% | [Yes/No] p-value: X.XX |

**Decision:** [Continue / Deploy Winner / Stop Test]
**Next Steps:** [Actions to take]

### Blockers and Dependencies

1. [Blocker 1]
   - Impact: [Description]
   - Owner: [Name]
   - ETA for resolution: [Date]

---

## 4. Backlog Prioritization

### New Optimization Opportunities

| ID | Title | Category | Impact | Effort | ROI | Priority |
|----|-------|----------|--------|--------|-----|----------|
| | | | X/10 | X/5 | XX | PX |

### Reprioritization Needed

- [Optimization ID]: Reason for reprioritization
- [Optimization ID]: Reason for reprioritization

### Next Sprint Commitments

1. [Optimization ID] - [Title] - Owner: [Name]
2. [Optimization ID] - [Title] - Owner: [Name]
3. [Optimization ID] - [Title] - Owner: [Name]

---

## 5. Capacity Planning

### Resource Utilization Trends

| Resource | Current | 7-Day Trend | Forecast (30d) | Action Needed |
|----------|---------|-------------|----------------|---------------|
| CPU Avg | XX% | +/-Y% | ZZ% | [Yes/No] |
| Memory Avg | XX% | +/-Y% | ZZ% | [Yes/No] |
| Storage | XX GB | +Y GB/week | ZZ GB | [Yes/No] |
| Network | XX Mbps | +/-Y% | ZZ Mbps | [Yes/No] |

### Scaling Events This Week

- [Date/Time]: Scaled [up/down] [component] from X to Y instances
  - Reason: [Why]
  - Impact: [Performance improvement or cost savings]

### Upcoming Capacity Needs

- [Date]: [Event or milestone requiring capacity planning]
- [Date]: [Event or milestone requiring capacity planning]

---

## 6. Cost Review

### This Week's Costs

| Category | This Week | Last Week | Change | Notes |
|----------|-----------|-----------|--------|-------|
| Compute | $XXX | $YYY | +/-Z% | |
| Database | $XXX | $YYY | +/-Z% | |
| Cache | $XXX | $YYY | +/-Z% | |
| Network | $XXX | $YYY | +/-Z% | |
| Monitoring | $XXX | $YYY | +/-Z% | |
| **Total** | **$XXX** | **$YYY** | **+/-Z%** | |

**Cost per Million Requests:** $XXX (Last week: $YYY, Change: +/-Z%)

### Cost Optimization Opportunities

1. [Opportunity 1]: Estimated savings $XXX/month
2. [Opportunity 2]: Estimated savings $XXX/month

---

## 7. Action Items

### From This Meeting

| Action Item | Owner | Due Date | Priority |
|-------------|-------|----------|----------|
| | | | |

### Follow-up from Previous Meetings

| Action Item | Owner | Status | Notes |
|-------------|-------|--------|-------|
| | | [Done/In Progress/Blocked] | |

---

## 8. Notes and Discussions

### Key Discussion Points

- [Topic 1]: [Summary of discussion and decisions]
- [Topic 2]: [Summary of discussion and decisions]

### Deep-Dive Sessions Needed

- [Topic]: Scheduled for [Date/Time], Attendees: [List]

---

## 9. Next Week's Focus

### Priorities

1. [Priority 1]
2. [Priority 2]
3. [Priority 3]

### Risks and Concerns

- [Risk 1]: Mitigation strategy: [Description]
- [Risk 2]: Mitigation strategy: [Description]

---

**Next Review:** [Date and Time]
**Meeting Notes Link:** [Link to detailed notes if applicable]
