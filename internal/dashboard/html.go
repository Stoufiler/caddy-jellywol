package dashboard

const statusPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>JellyWolProxy Control Panel</title>
    <style>
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
            color: #e2e8f0;
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1400px;
            margin: 0 auto;
        }
        h1 {
            font-size: 2em;
            margin-bottom: 30px;
            color: #f1f5f9;
            font-weight: 600;
        }
        .section {
            margin-bottom: 25px;
            background: rgba(255,255,255,0.05);
            border-radius: 8px;
            padding: 20px;
            border: 1px solid rgba(255,255,255,0.1);
        }
        .section-title {
            color: #38bdf8;
            font-size: 1.2em;
            margin-bottom: 15px;
            font-weight: 600;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
        }
        .metric {
            padding: 15px;
            background: rgba(15, 23, 42, 0.6);
            border-radius: 6px;
            border: 1px solid rgba(255,255,255,0.1);
        }
        .metric-label {
            font-size: 0.85em;
            color: #94a3b8;
            margin-bottom: 8px;
        }
        .metric-value {
            font-size: 1.6em;
            font-weight: 600;
            color: #f1f5f9;
        }
        .status-online { color: #22c55e; }
        .status-offline { color: #ef4444; }
        .status-waking { color: #f59e0b; }

        /* Sessions */
        .session-card {
            background: rgba(30, 41, 59, 0.5);
            border: 1px solid rgba(255,255,255,0.1);
            border-radius: 8px;
            padding: 16px;
            margin-bottom: 12px;
            transition: transform 0.2s, box-shadow 0.2s;
        }
        .session-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(0,0,0,0.3);
        }
        .session-user {
            font-size: 1.1em;
            color: #f1f5f9;
            margin-bottom: 8px;
            font-weight: 600;
        }
        .session-item {
            font-size: 1em;
            color: #38bdf8;
            margin-bottom: 8px;
        }
        .session-device {
            font-size: 0.9em;
            color: #94a3b8;
            margin-bottom: 10px;
        }
        .progress-bar {
            width: 100%;
            height: 8px;
            background: rgba(255,255,255,0.1);
            border-radius: 4px;
            overflow: hidden;
            margin-top: 10px;
        }
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #3b82f6, #06b6d4);
            transition: width 1s;
        }
        .playstate {
            display: inline-block;
            padding: 3px 10px;
            border-radius: 4px;
            font-size: 0.75em;
            font-weight: 600;
            margin-top: 8px;
        }
        .playing {
            background: rgba(34, 197, 94, 0.2);
            color: #22c55e;
            border: 1px solid #22c55e;
        }
        .paused {
            background: rgba(245, 158, 11, 0.2);
            color: #f59e0b;
            border: 1px solid #f59e0b;
        }

        /* Chart */
        .chart-container {
            background: rgba(15, 23, 42, 0.6);
            border: 1px solid rgba(255,255,255,0.1);
            border-radius: 8px;
            padding: 15px;
            margin-top: 15px;
        }
        canvas { width: 100% !important; height: 150px !important; }

        /* Logs */
        .logs {
            background: rgba(15, 23, 42, 0.8);
            border: 1px solid rgba(255,255,255,0.1);
            border-radius: 8px;
            padding: 15px;
            height: 300px;
            overflow-y: auto;
            font-family: 'Monaco', 'Menlo', monospace;
            font-size: 0.85em;
        }
        .log-entry {
            padding: 4px 0;
            border-bottom: 1px solid rgba(255,255,255,0.05);
        }
        .log-time { color: #94a3b8; }
        .log-error { color: #ef4444; }
        .log-warn { color: #f59e0b; }
        .log-info { color: #38bdf8; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📊 JellyWolProxy Dashboard</h1>

        <div class="section">
            <div class="section-title">🎬 Active Streaming Sessions</div>
            <div id="sessions">
                    <div style="color: #888; font-style: italic;">No active sessions...</div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">📈 Server Metrics</div>
                <div class="grid">
                    <div class="metric">
                        <div class="metric-label">Jellyfin Status</div>
                        <div class="metric-value" id="serverState">CHECKING...</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Uptime</div>
                        <div class="metric-value" id="uptime">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Total Requests</div>
                        <div class="metric-value" id="totalRequests">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Cache Hit Rate</div>
                        <div class="metric-value" id="cacheHitRate">0%</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Wake-up Count</div>
                        <div class="metric-value" id="wakeUpCount">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Avg Wake Time</div>
                        <div class="metric-value" id="avgWakeUpTime">-</div>
                    </div>
                </div>

                <div class="chart-container">
                    <div style="color: #38bdf8; margin-bottom: 10px; font-weight: 600;">Network Bandwidth</div>
                    <canvas id="bandwidthChart"></canvas>
                    <div style="display: flex; justify-content: space-around; margin-top: 10px; font-size: 0.9em;">
                        <div>↓ <span id="bandwidthIn" style="color: #06b6d4;">0 B/s</span></div>
                        <div>↑ <span id="bandwidthOut" style="color: #3b82f6;">0 B/s</span></div>
                        <div>Total In: <span id="totalBytesIn">0 B</span></div>
                        <div>Total Out: <span id="totalBytesOut">0 B</span></div>
                    </div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">💻 System Information</div>
                <div class="grid">
                    <div class="metric">
                        <div class="metric-label">Hostname</div>
                        <div class="metric-value" style="font-size: 1.1em;" id="hostname">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">OS / Arch</div>
                        <div class="metric-value" style="font-size: 1.1em;" id="osArch">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">CPU Cores</div>
                        <div class="metric-value" id="numCPU">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Memory (Alloc/Sys)</div>
                        <div class="metric-value" style="font-size: 1em;" id="memory">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">Goroutines</div>
                        <div class="metric-value" id="goroutines">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">GC Runs</div>
                        <div class="metric-value" id="gcCount">-</div>
                    </div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">📝 Live System Logs</div>
                <div class="logs" id="logs">
                    <div class="log-entry"><span class="log-time">[INIT]</span> Connecting to log stream...</div>
                </div>
            </div>

            <div style="text-align: center; margin-top: 20px; color: #94a3b8; font-size: 0.85em;">
                Auto-refresh every 2 seconds • Real-time monitoring
            </div>
        </div>

    <script>
        // Bandwidth chart
        const maxDataPoints = 60;
        const bandwidthInData = new Array(maxDataPoints).fill(0);
        const bandwidthOutData = new Array(maxDataPoints).fill(0);
        let maxBandwidth = 1024;

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }

        function formatBandwidth(bytesPerSec) {
            return formatBytes(bytesPerSec) + '/s';
        }

        function drawChart() {
            const canvas = document.getElementById('bandwidthChart');
            const ctx = canvas.getContext('2d');
            const dpr = window.devicePixelRatio || 1;

            canvas.width = canvas.offsetWidth * dpr;
            canvas.height = 150 * dpr;
            ctx.scale(dpr, dpr);

            const width = canvas.offsetWidth;
            const height = 150;
            const padding = 5;

            ctx.clearRect(0, 0, width, height);

            const currentMax = Math.max(...bandwidthInData, ...bandwidthOutData, 1024);
            maxBandwidth = Math.max(maxBandwidth * 0.95, currentMax * 1.2);

            // Grid
            ctx.strokeStyle = 'rgba(255,255,255,0.1)';
            ctx.lineWidth = 1;
            for (let i = 0; i < 4; i++) {
                const y = padding + (height - padding * 2) * i / 3;
                ctx.beginPath();
                ctx.moveTo(0, y);
                ctx.lineTo(width, y);
                ctx.stroke();
            }

            // IN (cyan/blue gradient)
            ctx.fillStyle = 'rgba(6,182,212,0.2)';
            ctx.strokeStyle = '#06b6d4';
            ctx.lineWidth = 2;
            ctx.beginPath();
            ctx.moveTo(0, height - padding);
            for (let i = 0; i < bandwidthInData.length; i++) {
                const x = (i / (bandwidthInData.length - 1)) * width;
                const y = height - padding - (bandwidthInData[i] / maxBandwidth) * (height - padding * 2);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.stroke();
            ctx.lineTo(width, height - padding);
            ctx.lineTo(0, height - padding);
            ctx.fill();

            // OUT (blue)
            ctx.fillStyle = 'rgba(59,130,246,0.2)';
            ctx.strokeStyle = '#3b82f6';
            ctx.lineWidth = 2;
            ctx.beginPath();
            for (let i = 0; i < bandwidthOutData.length; i++) {
                const x = (i / (bandwidthOutData.length - 1)) * width;
                const y = height - padding - (bandwidthOutData[i] / maxBandwidth) * (height - padding * 2);
                if (i === 0) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.stroke();
            ctx.lineTo(width, height - padding);
            ctx.lineTo(0, height - padding);
            ctx.fill();
        }

        function updateSessions(sessions) {
            const container = document.getElementById('sessions');
            if (!sessions || sessions.length === 0) {
                container.innerHTML = '<div style="color: #888; font-style: italic;">No active sessions...</div>';
                return;
            }

            const activeSessions = sessions.filter(s => s.NowPlayingItem);
            if (activeSessions.length === 0) {
                container.innerHTML = '<div style="color: #888; font-style: italic;">No active playback sessions...</div>';
                return;
            }

            container.innerHTML = activeSessions.map(session => {
                const item = session.NowPlayingItem;
                const title = item.Type === 'Episode' && item.SeriesName
                    ? item.SeriesName + ' - S' + String(item.ParentIndexNumber || 0).padStart(2, '0') + 'E' + String(item.IndexNumber || 0).padStart(2, '0') + ' - ' + item.Name
                    : item.Name;

                const progress = item.RunTimeTicks > 0
                    ? (session.PlayState.PositionTicks / item.RunTimeTicks * 100).toFixed(1)
                    : 0;

                const playstate = session.PlayState.IsPaused ? 'PAUSED' : 'PLAYING';
                const stateClass = session.PlayState.IsPaused ? 'paused' : 'playing';

                return '<div class="session-card">' +
                    '<div class="session-user">👤 ' + session.UserName + '</div>' +
                    '<div class="session-item">📺 ' + title + '</div>' +
                    '<div class="session-device">💻 ' + session.DeviceName + ' (' + session.Client + ')</div>' +
                    '<div class="playstate ' + stateClass + '">' + playstate + '</div>' +
                    '<div class="progress-bar">' +
                        '<div class="progress-fill" style="width: ' + progress + '%"></div>' +
                    '</div>' +
                    '<div style="margin-top: 5px; font-size: 0.9em; color: #888;">' + progress + '% complete</div>' +
                '</div>';
            }).join('');
        }

        function updateStatus() {
            fetch('/status/api')
                .then(r => r.json())
                .then(data => {
                    const stateEl = document.getElementById('serverState');
                    stateEl.textContent = data.serverState.toUpperCase();
                    stateEl.className = 'metric-value status-' + data.serverState;

                    document.getElementById('uptime').textContent = data.uptime;
                    document.getElementById('totalRequests').textContent = data.totalRequests.toLocaleString();
                    document.getElementById('cacheHitRate').textContent = data.cacheHitRate.toFixed(1) + '%';
                    document.getElementById('wakeUpCount').textContent = data.wakeUpCount;
                    document.getElementById('avgWakeUpTime').textContent = data.avgWakeUpTimeSeconds > 0
                        ? data.avgWakeUpTimeSeconds.toFixed(1) + 's' : '-';

                    document.getElementById('bandwidthIn').textContent = formatBandwidth(data.bandwidthIn);
                    document.getElementById('bandwidthOut').textContent = formatBandwidth(data.bandwidthOut);
                    document.getElementById('totalBytesIn').textContent = formatBytes(data.bytesIn);
                    document.getElementById('totalBytesOut').textContent = formatBytes(data.bytesOut);

                    bandwidthInData.shift();
                    bandwidthInData.push(data.bandwidthIn);
                    bandwidthOutData.shift();
                    bandwidthOutData.push(data.bandwidthOut);
                    drawChart();

                    document.getElementById('hostname').textContent = data.system.hostname;
                    document.getElementById('osArch').textContent = data.system.os + ' / ' + data.system.arch;
                    document.getElementById('numCPU').textContent = data.system.numCpu;
                    document.getElementById('memory').textContent = data.system.memAllocMB.toFixed(1) + ' / ' + data.system.memSysMB.toFixed(1) + ' MB';
                    document.getElementById('goroutines').textContent = data.system.numGoroutines.toLocaleString();
                    document.getElementById('gcCount').textContent = data.system.gcCount.toLocaleString();

                    updateSessions(data.sessions);
                })
                .catch(err => console.error('Failed to fetch status:', err));
        }

        updateStatus();
        setInterval(updateStatus, 2000);
        window.addEventListener('resize', drawChart);

        // Logs
        const logsEl = document.getElementById('logs');
        const evtSource = new EventSource('/status/logs');
        evtSource.onmessage = function(e) {
            const entry = document.createElement('div');
            entry.className = 'log-entry';
            const data = JSON.parse(e.data);
            let levelClass = '';
            if (data.level === 'warning') levelClass = 'log-warn';
            if (data.level === 'error') levelClass = 'log-error';
            entry.innerHTML = '<span class="log-time">[' + data.time + ']</span> ' +
                '<span class="' + levelClass + '">' + data.message + '</span>';
            logsEl.appendChild(entry);
            logsEl.scrollTop = logsEl.scrollHeight;
            while (logsEl.children.length > 100) {
                logsEl.removeChild(logsEl.firstChild);
            }
        };
    </script>
</body>
</html>`
