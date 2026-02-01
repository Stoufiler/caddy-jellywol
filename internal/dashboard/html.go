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
            font-family: 'Courier New', monospace;
            background: #0a0e27;
            color: #00ff41;
            min-height: 100vh;
            padding: 20px;
            background-image:
                repeating-linear-gradient(0deg, rgba(0,255,65,0.03) 0px, transparent 1px, transparent 2px, rgba(0,255,65,0.03) 3px),
                repeating-linear-gradient(90deg, rgba(0,255,65,0.03) 0px, transparent 1px, transparent 2px, rgba(0,255,65,0.03) 3px);
        }
        .terminal {
            max-width: 1800px;
            margin: 0 auto;
            border: 2px solid #00ff41;
            box-shadow: 0 0 20px rgba(0,255,65,0.3);
            background: rgba(10,14,39,0.95);
        }
        .terminal-header {
            background: #00ff41;
            color: #0a0e27;
            padding: 8px 16px;
            font-weight: bold;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .terminal-body { padding: 20px; }
        .prompt { color: #00ff41; }
        .prompt::before { content: '> '; }
        h1 {
            font-size: 2em;
            margin-bottom: 20px;
            text-shadow: 0 0 10px rgba(0,255,65,0.8);
            animation: flicker 3s infinite alternate;
        }
        @keyframes flicker {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.95; }
        }
        .section {
            margin-bottom: 30px;
            border: 1px solid rgba(0,255,65,0.3);
            padding: 15px;
            background: rgba(0,255,65,0.02);
        }
        .section-title {
            color: #00d9ff;
            font-size: 1.3em;
            margin-bottom: 15px;
            border-bottom: 1px solid rgba(0,255,65,0.3);
            padding-bottom: 8px;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 15px;
        }
        .metric {
            padding: 12px;
            border: 1px solid rgba(0,255,65,0.2);
            background: rgba(0,20,40,0.6);
        }
        .metric-label {
            font-size: 0.85em;
            color: #00d9ff;
            margin-bottom: 5px;
        }
        .metric-value {
            font-size: 1.4em;
            font-weight: bold;
        }
        .status-online { color: #00ff41; }
        .status-offline { color: #ff0055; }
        .status-waking { color: #ffaa00; }

        /* Sessions */
        .session-card {
            border: 1px solid rgba(0,255,65,0.4);
            padding: 15px;
            margin-bottom: 12px;
            background: rgba(0,40,20,0.3);
            position: relative;
            overflow: hidden;
        }
        .session-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: linear-gradient(90deg, transparent, #00ff41, transparent);
            animation: scan 2s linear infinite;
        }
        @keyframes scan {
            0% { transform: translateX(-100%); }
            100% { transform: translateX(100%); }
        }
        .session-user {
            font-size: 1.2em;
            color: #00ff41;
            margin-bottom: 8px;
        }
        .session-item {
            font-size: 1.1em;
            color: #00d9ff;
            margin-bottom: 8px;
        }
        .session-device {
            font-size: 0.9em;
            color: #888;
            margin-bottom: 8px;
        }
        .progress-bar {
            width: 100%;
            height: 6px;
            background: rgba(0,255,65,0.2);
            border: 1px solid rgba(0,255,65,0.3);
            position: relative;
            margin-top: 8px;
        }
        .progress-fill {
            height: 100%;
            background: #00ff41;
            transition: width 1s;
            box-shadow: 0 0 10px rgba(0,255,65,0.8);
        }
        .playstate {
            display: inline-block;
            padding: 2px 8px;
            border: 1px solid;
            font-size: 0.8em;
            margin-top: 5px;
        }
        .playing { border-color: #00ff41; color: #00ff41; }
        .paused { border-color: #ffaa00; color: #ffaa00; }

        /* Chart */
        .chart-container {
            background: rgba(0,20,40,0.6);
            border: 1px solid rgba(0,255,65,0.2);
            padding: 15px;
            margin-top: 15px;
        }
        canvas { width: 100% !important; height: 150px !important; }

        /* Logs */
        .logs {
            background: #000;
            border: 1px solid rgba(0,255,65,0.3);
            padding: 12px;
            height: 300px;
            overflow-y: auto;
            font-size: 0.9em;
        }
        .log-entry {
            padding: 4px 0;
            border-bottom: 1px solid rgba(0,255,65,0.1);
        }
        .log-time { color: #00d9ff; }
        .log-error { color: #ff0055; }
        .log-warn { color: #ffaa00; }

        .blink { animation: blink 1s step-start infinite; }
        @keyframes blink {
            50% { opacity: 0; }
        }
    </style>
</head>
<body>
    <div class="terminal">
        <div class="terminal-header">
            <span>JELLYWOLPROXY CONTROL PANEL v1.0</span>
            <span id="clock"></span>
        </div>
        <div class="terminal-body">
            <h1>┌─[ SYSTEM STATUS ]</h1>

            <div class="section">
                <div class="section-title">▸ ACTIVE STREAMING SESSIONS</div>
                <div id="sessions">
                    <div style="color: #888; font-style: italic;">No active sessions...</div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">▸ SERVER METRICS</div>
                <div class="grid">
                    <div class="metric">
                        <div class="metric-label">JELLYFIN STATUS</div>
                        <div class="metric-value" id="serverState">CHECKING<span class="blink">_</span></div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">UPTIME</div>
                        <div class="metric-value" id="uptime">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">TOTAL REQUESTS</div>
                        <div class="metric-value" id="totalRequests">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">CACHE HIT RATE</div>
                        <div class="metric-value" id="cacheHitRate">0%</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">WAKE-UP COUNT</div>
                        <div class="metric-value" id="wakeUpCount">0</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">AVG WAKE TIME</div>
                        <div class="metric-value" id="avgWakeUpTime">-</div>
                    </div>
                </div>

                <div class="chart-container">
                    <div style="color: #00d9ff; margin-bottom: 10px;">NETWORK BANDWIDTH (IN/OUT)</div>
                    <canvas id="bandwidthChart"></canvas>
                    <div style="display: flex; justify-content: space-around; margin-top: 10px; font-size: 0.9em;">
                        <div>↓ <span id="bandwidthIn">0 B/s</span></div>
                        <div>↑ <span id="bandwidthOut">0 B/s</span></div>
                        <div>TOTAL IN: <span id="totalBytesIn">0 B</span></div>
                        <div>TOTAL OUT: <span id="totalBytesOut">0 B</span></div>
                    </div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">▸ SYSTEM INFORMATION</div>
                <div class="grid">
                    <div class="metric">
                        <div class="metric-label">HOSTNAME</div>
                        <div class="metric-value" style="font-size: 1.1em;" id="hostname">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">OS / ARCH</div>
                        <div class="metric-value" style="font-size: 1.1em;" id="osArch">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">CPU CORES</div>
                        <div class="metric-value" id="numCPU">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">MEMORY (ALLOC/SYS)</div>
                        <div class="metric-value" style="font-size: 1em;" id="memory">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">GOROUTINES</div>
                        <div class="metric-value" id="goroutines">-</div>
                    </div>
                    <div class="metric">
                        <div class="metric-label">GC RUNS</div>
                        <div class="metric-value" id="gcCount">-</div>
                    </div>
                </div>
            </div>

            <div class="section">
                <div class="section-title">▸ LIVE SYSTEM LOGS</div>
                <div class="logs" id="logs">
                    <div class="log-entry"><span class="log-time">[INIT]</span> Connecting to log stream...</div>
                </div>
            </div>

            <div style="text-align: center; margin-top: 20px; color: #00d9ff; font-size: 0.85em;">
                <span class="blink">█</span> AUTO-REFRESH: 2s | REAL-TIME MONITORING ACTIVE
            </div>
        </div>
    </div>

    <script>
        // Clock
        function updateClock() {
            const now = new Date();
            document.getElementById('clock').textContent = now.toLocaleTimeString('en-US', { hour12: false });
        }
        setInterval(updateClock, 1000);
        updateClock();

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
            ctx.strokeStyle = 'rgba(0,255,65,0.1)';
            ctx.lineWidth = 1;
            for (let i = 0; i < 4; i++) {
                const y = padding + (height - padding * 2) * i / 3;
                ctx.beginPath();
                ctx.moveTo(0, y);
                ctx.lineTo(width, y);
                ctx.stroke();
            }

            // IN (green)
            ctx.fillStyle = 'rgba(0,255,65,0.2)';
            ctx.strokeStyle = '#00ff41';
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

            // OUT (cyan)
            ctx.fillStyle = 'rgba(0,217,255,0.2)';
            ctx.strokeStyle = '#00d9ff';
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
