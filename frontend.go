// Package main provides the entry point, webhook handlers, and utility functions for the DaVinci GitHub Bot.
package main

import "fmt"

// GetDashboardHTML returns the FGO-themed product landing page.
func GetDashboardHTML(version, appName, avatarURL, appHTMLURL string, activeRepos int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ScathachGrip/davinci</title>
    <link href="https://fonts.googleapis.com/css2?family=Cinzel:wght@500;700;800&family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            /* Default Dark Theme Variables */
            --bg: #07090e;
            --panel-bg: rgba(14, 18, 30, 0.95);
            --gold: #d4af37;
            --gold-light: #f3e5ab;
            --text: #e2e8f0;
            --text-muted: #94a3b8;
            --border: #1e293b;
            --border-gold: rgba(212, 175, 55, 0.4);
            --rarity: #f59e0b;
            --radial-color: rgba(212, 175, 55, 0.05);
            --card-gradient: linear-gradient(180deg, #111827 0%%, #030712 100%%);
            --card-illustration-bg: #07090e;
            --meta-bg: rgba(255, 255, 255, 0.02);
            --meta-border: rgba(255, 255, 255, 0.04);
            --btn-hover-bg: rgba(212, 175, 55, 0.1);
        }

        body.light-theme {
            /* Light Theme Variables */
            --bg: #f8fafc;
            --panel-bg: rgba(255, 255, 255, 0.95);
            --gold: #b58d16;
            --gold-light: #7c5e00;
            --text: #0f172a;
            --text-muted: #64748b;
            --border: #cbd5e1;
            --border-gold: rgba(181, 141, 22, 0.3);
            --rarity: #d97706;
            --radial-color: rgba(181, 141, 22, 0.03);
            --card-gradient: linear-gradient(180deg, #f8fafc 0%%, #f1f5f9 100%%);
            --card-illustration-bg: #ffffff;
            --meta-bg: rgba(0, 0, 0, 0.02);
            --meta-border: rgba(0, 0, 0, 0.05);
            --btn-hover-bg: rgba(181, 141, 22, 0.1);
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            background-color: var(--bg);
            background-image: 
                radial-gradient(at 10%% 10%%, var(--radial-color) 0px, transparent 40%%),
                radial-gradient(at 90%% 90%%, var(--radial-color) 0px, transparent 40%%);
            color: var(--text);
            font-family: 'Plus Jakarta Sans', sans-serif;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 40px 20px;
            transition: background-color 0.3s ease, color 0.3s ease;
        }

        .container {
            width: 100%%;
            max-width: 900px;
            display: grid;
            grid-template-columns: 320px 1fr;
            gap: 40px;
            align-items: center;
        }

        @media (max-width: 768px) {
            .container {
                grid-template-columns: 1fr;
                gap: 32px;
            }
        }

        @keyframes slideUp {
            from {
                opacity: 0;
                transform: translateY(30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        /* FGO Servant Card Styling */
        .card-container {
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            opacity: 0;
            animation: slideUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
            animation-delay: 0.1s;
        }

        .active-repos-badge {
            margin-top: 20px;
            font-family: 'Cinzel', serif;
            font-size: 11px;
            color: var(--gold-light);
            border: 1px solid var(--border-gold);
            background-color: var(--meta-bg);
            padding: 6px 16px;
            border-radius: 20px;
            letter-spacing: 0.08em;
            text-transform: uppercase;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
            transition: color 0.3s ease, border-color 0.3s ease;
        }

        .servant-card {
            width: 280px;
            height: 440px;
            background: var(--card-gradient);
            border: 4px solid var(--gold);
            border-radius: 12px;
            box-shadow: 
                0 0 20px rgba(212, 175, 55, 0.15),
                0 25px 50px -12px rgba(0, 0, 0, 0.5);
            position: relative;
            padding: 16px;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            overflow: hidden;
            transition: background 0.3s ease;
        }

        /* Gold frame patterns */
        .servant-card::before {
            content: '';
            position: absolute;
            top: 4px; left: 4px; right: 4px; bottom: 4px;
            border: 1px solid rgba(212, 175, 55, 0.3);
            border-radius: 6px;
            pointer-events: none;
        }

        .card-header {
            text-align: center;
            z-index: 2;
        }

        .class-title {
            font-family: 'Cinzel', serif;
            font-size: 11px;
            font-weight: 700;
            color: var(--gold-light);
            letter-spacing: 0.15em;
            text-transform: uppercase;
        }

        .stars {
            color: var(--rarity);
            font-size: 14px;
            margin-top: 4px;
            letter-spacing: 2px;
        }

        .card-illustration {
            width: 140px;
            height: 140px;
            margin: 24px auto;
            border-radius: 50%%;
            border: 3px solid var(--gold);
            overflow: hidden;
            position: relative;
            box-shadow: 0 0 15px rgba(212, 175, 55, 0.15);
            z-index: 2;
            background-color: var(--card-illustration-bg);
            transition: background-color 0.3s ease;
        }

        .card-illustration img {
            width: 100%%;
            height: 100%%;
            object-fit: cover;
        }

        .card-footer {
            text-align: center;
            z-index: 2;
            border-top: 1px solid rgba(212, 175, 55, 0.2);
            padding-top: 16px;
        }

        .servant-name {
            font-family: 'Cinzel', serif;
            font-size: 18px;
            font-weight: 700;
            color: var(--text);
            letter-spacing: -0.02em;
            margin-bottom: 4px;
            transition: color 0.3s ease;
        }

        .servant-title {
            font-size: 11px;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.05em;
            font-weight: 600;
        }

        /* Right Side Content UI */
        .info-panel {
            background-color: var(--panel-bg);
            border: 2px solid var(--border-gold);
            border-radius: 16px;
            padding: 32px;
            box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.4);
            position: relative;
            opacity: 0;
            animation: slideUp 0.8s cubic-bezier(0.16, 1, 0.3, 1) forwards;
            animation-delay: 0.3s;
            transition: background-color 0.3s ease, border-color 0.3s ease;
        }

        /* Top right corner banner style decoration */
        .info-panel::before {
            content: '';
            position: absolute;
            top: 0; left: 0;
            width: 16px; height: 16px;
            border-top: 2px solid var(--gold);
            border-left: 2px solid var(--gold);
        }
        .info-panel::after {
            content: '';
            position: absolute;
            bottom: 0; right: 0;
            width: 16px; height: 16px;
            border-bottom: 2px solid var(--gold);
            border-right: 2px solid var(--gold);
        }

        h2.panel-title {
            font-family: 'Cinzel', serif;
            font-size: 28px;
            color: var(--text);
            margin-bottom: 12px;
            letter-spacing: -0.03em;
            transition: color 0.3s ease;
        }

        .panel-desc {
            color: var(--text-muted);
            font-size: 14px;
            line-height: 1.6;
            margin-bottom: 28px;
        }

        /* Skills block styling */
        .skills-container {
            display: flex;
            flex-direction: column;
            gap: 20px;
            margin-bottom: 32px;
        }

        .skill-row {
            display: grid;
            grid-template-columns: 44px 1fr;
            gap: 16px;
            align-items: flex-start;
        }

        .skill-icon {
            width: 44px;
            height: 44px;
            border: 2px solid var(--gold);
            border-radius: 8px;
            background-color: var(--card-illustration-bg);
            display: flex;
            align-items: center;
            justify-content: center;
            font-family: 'Cinzel', serif;
            font-weight: 700;
            color: var(--gold-light);
            font-size: 14px;
            box-shadow: 0 4px 10px rgba(0,0,0,0.1);
            transition: background-color 0.3s ease;
        }

        .skill-info {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }

        .skill-name {
            font-family: 'Cinzel', serif;
            font-size: 14px;
            font-weight: 700;
            color: var(--text);
            transition: color 0.3s ease;
        }

        .skill-desc {
            color: var(--text-muted);
            font-size: 13px;
            line-height: 1.5;
        }

        /* Action Buttons */
        .actions-row {
            display: flex;
            gap: 16px;
            border-top: 1px solid var(--border);
            padding-top: 24px;
            transition: border-color 0.3s ease;
        }

        .btn-gold {
            font-family: 'Cinzel', serif;
            font-size: 12px;
            font-weight: 700;
            color: var(--gold-light);
            background-color: transparent;
            border: 1px solid var(--gold);
            padding: 10px 24px;
            border-radius: 4px;
            text-decoration: none;
            text-align: center;
            transition: all 0.2s ease;
            letter-spacing: 0.1em;
            text-transform: uppercase;
        }

        .btn-gold:hover {
            background-color: var(--btn-hover-bg);
            color: var(--text);
            box-shadow: 0 0 10px var(--border-gold);
        }

        /* Meta table info */
        .meta-table {
            display: grid;
            grid-template-columns: 1fr 1fr 1fr;
            gap: 16px;
            margin-bottom: 24px;
            background: var(--meta-bg);
            border: 1px solid var(--meta-border);
            border-radius: 8px;
            padding: 12px 16px;
            font-size: 12px;
            transition: background 0.3s ease, border-color 0.3s ease;
        }

        .meta-col {
            display: flex;
            flex-direction: column;
            gap: 4px;
        }

        .meta-label {
            color: var(--text-muted);
            text-transform: uppercase;
            font-weight: 600;
            font-size: 10px;
            letter-spacing: 0.05em;
        }

        .meta-value {
            font-family: ui-monospace, SFMono-Regular, SF Mono, Menlo, Consolas, Liberation Mono, monospace;
            color: var(--text);
            font-weight: 600;
            transition: color 0.3s ease;
        }

        /* Floating Theme Toggle */
        .theme-toggle {
            position: fixed;
            bottom: 24px;
            right: 24px;
            width: 40px;
            height: 40px;
            border-radius: 50%%;
            border: 2px solid var(--gold);
            background-color: var(--panel-bg);
            color: var(--gold-light);
            cursor: pointer;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 4px 10px rgba(0,0,0,0.2);
            transition: all 0.2s ease;
            z-index: 100;
        }

        .theme-toggle:hover {
            background-color: var(--btn-hover-bg);
            color: var(--text);
            box-shadow: 0 0 10px var(--border-gold);
        }

        .toggle-icon {
            width: 16px;
            height: 16px;
            stroke-width: 2;
        }
    </style>
</head>
<body>
    <!-- Theme Toggle -->
    <button class="theme-toggle" id="theme-toggle" aria-label="Toggle theme">
        <!-- Sun icon (for dark mode -> switch to light) -->
        <svg id="sun-icon" class="toggle-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="5"></circle>
            <line x1="12" y1="1" x2="12" y2="3"></line>
            <line x1="12" y1="21" x2="12" y2="23"></line>
            <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
            <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
            <line x1="1" y1="12" x2="3" y2="12"></line>
            <line x1="21" y1="12" x2="23" y2="12"></line>
            <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
            <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
        </svg>
        <!-- Moon icon (for light mode -> switch to dark) -->
        <svg id="moon-icon" class="toggle-icon" style="display: none;" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
        </svg>
    </button>

    <div class="container">
        <!-- Left Side: FGO Servant Card -->
        <div class="card-container">
            <div class="servant-card">
                <div class="card-header">
                    <div class="class-title">ScathachGrip/davinci</div>
                    <div class="stars">★★★★★</div>
                </div>

                <div class="card-illustration">
                    <img src="%[3]s" alt="App Avatar">
                </div>

                <div class="card-footer">
                    <div class="servant-name">%[2]s</div>
                    <div class="servant-title">workflow automator</div>
                </div>
            </div>
            <div class="active-repos-badge">Listening to %[5]d Repositories👀</div>
        </div>

        <!-- Right Side: Clean Game-like UI Panel -->
        <div class="info-panel">
            <h2 class="panel-title">Universal Automation</h2>
            <p class="panel-desc">Experimental github bots to automate PR and Issues workflows</p>

            <div class="meta-table">
                <div class="meta-col">
                    <span class="meta-label">App ID</span>
                    <span class="meta-value">243623</span>
                </div>
                <div class="meta-col">
                    <span class="meta-label">Version</span>
                    <span class="meta-value">v%[1]s</span>
                </div>
                <div class="meta-col">
                    <span class="meta-label">Runtime</span>
                    <span class="meta-value">Go 1.26</span>
                </div>
            </div>

            <!-- Active Skills (Features) -->
            <div class="skills-container">
                <div class="skill-row">
                    <div class="skill-icon">A</div>
                    <div class="skill-info">
                        <span class="skill-name">Golden Rule (Triage) A</span>
                        <span class="skill-desc">Adds triage label and comments.</span>
                    </div>
                </div>
                <div class="skill-row">
                    <div class="skill-icon">EX</div>
                    <div class="skill-info">
                        <span class="skill-name">Pioneer of the Stars (PR Lifecycle) EX</span>
                        <span class="skill-desc">Tags pull requests with pr:pending, and automatically cleans up the label on merge/close.</span>
                    </div>
                </div>
                <div class="skill-row">
                    <div class="skill-icon">EX</div>
                    <div class="skill-info">
                        <span class="skill-name">Universal Man (LGTM Auto-Merge) EX</span>
                        <span class="skill-desc">Merges PRs instantly on 'lgtm' comment by authorized maintainers, preserving the repository's git history merge technique.</span>
                    </div>
                </div>
            </div>

            <!-- Actions -->
            <div class="actions-row">
                <a href="%[4]s" target="_blank" class="btn-gold">Install App</a>
                <a href="https://github.com/ScathachGrip/davinci" target="_blank" class="btn-gold">Contribute</a>
            </div>
        </div>
    </div>

    <!-- Theme Toggle Script -->
    <script>
        const toggleBtn = document.getElementById('theme-toggle');
        const sunIcon = document.getElementById('sun-icon');
        const moonIcon = document.getElementById('moon-icon');

        // Check and apply stored preference
        const currentTheme = localStorage.getItem('theme');
        if (currentTheme === 'light') {
            document.body.classList.add('light-theme');
            sunIcon.style.display = 'none';
            moonIcon.style.display = 'block';
        }

        toggleBtn.addEventListener('click', () => {
            document.body.classList.toggle('light-theme');
            const isLight = document.body.classList.contains('light-theme');
            localStorage.setItem('theme', isLight ? 'light' : 'dark');

            if (isLight) {
                sunIcon.style.display = 'none';
                moonIcon.style.display = 'block';
            } else {
                sunIcon.style.display = 'block';
                moonIcon.style.display = 'none';
            }
        });
    </script>
</body>
</html>`, version, appName, avatarURL, appHTMLURL, activeRepos)
}
