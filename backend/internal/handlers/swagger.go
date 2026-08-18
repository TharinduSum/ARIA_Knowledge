package handlers

import (
	_ "embed"
	"net/http"
)

//go:embed swagger.json
var swaggerJSON []byte

// SwaggerJSONHandler serves the OpenAPI specification JSON file.
func SwaggerJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Write(swaggerJSON)
}

// SwaggerUIHandler serves the customized dark-themed Swagger UI HTML page.
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerHTML))
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ARIA Knowledge - API Documentation</title>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <!-- Fonts -->
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700&display=swap" rel="stylesheet">
  
  <!-- Swagger UI default styles -->
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  
  <!-- Beautiful Dark Theme for Swagger UI -->
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/gh/Itz-fork/Fastapi-Swagger-UI-Dark/assets/swagger_ui_dark.min.css" />
  
  <style>
    /* Premium overrides */
    body {
      margin: 0;
      background-color: #0b0f19 !important;
      font-family: 'Outfit', sans-serif !important;
    }
    
    /* Top Bar design */
    .aria-header {
      background: linear-gradient(135deg, #0f172a 0%, #1e1b4b 100%);
      padding: 1.5rem 2rem;
      border-bottom: 1px solid rgba(99, 102, 241, 0.2);
      display: flex;
      align-items: center;
      justify-content: space-between;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.3);
    }
    
    .aria-logo-container {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }
    
    .aria-logo-icon {
      width: 2.5rem;
      height: 2.5rem;
      background: linear-gradient(135deg, #6366f1 0%, #a855f7 100%);
      border-radius: 0.5rem;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 700;
      color: #ffffff;
      font-size: 1.25rem;
      box-shadow: 0 0 15px rgba(99, 102, 241, 0.5);
      animation: pulse 2s infinite alternate;
    }
    
    @keyframes pulse {
      0% { transform: scale(1); box-shadow: 0 0 10px rgba(99, 102, 241, 0.4); }
      100% { transform: scale(1.05); box-shadow: 0 0 20px rgba(168, 85, 247, 0.6); }
    }
    
    .aria-title {
      font-size: 1.5rem;
      font-weight: 700;
      color: #ffffff;
      letter-spacing: -0.025em;
      margin: 0;
    }
    
    .aria-subtitle {
      font-size: 0.875rem;
      color: #a5b4fc;
      margin: 0;
      opacity: 0.8;
    }

    .status-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.5rem;
      background: rgba(16, 185, 129, 0.1);
      border: 1px solid rgba(16, 185, 129, 0.2);
      padding: 0.375rem 0.75rem;
      border-radius: 9999px;
      color: #34d399;
      font-size: 0.875rem;
      font-weight: 500;
    }
    
    .status-dot {
      width: 8px;
      height: 8px;
      background-color: #10b981;
      border-radius: 50%;
      box-shadow: 0 0 8px #10b981;
      animation: blink 1.5s infinite;
    }
    
    @keyframes blink {
      0%, 100% { opacity: 0.5; }
      50% { opacity: 1; }
    }
    
    /* Swagger UI Styles overrides */
    .swagger-ui {
      background-color: #0b0f19 !important;
      color: #cbd5e1 !important;
    }
    
    .swagger-ui .info {
      margin: 20px 0 !important;
    }
    
    .swagger-ui .info .title {
      color: #ffffff !important;
      font-family: 'Outfit', sans-serif !important;
      font-size: 2.25rem !important;
      font-weight: 700 !important;
      margin: 0 !important;
    }
    
    .swagger-ui .info p, 
    .swagger-ui .info li, 
    .swagger-ui .info table {
      color: #94a3b8 !important;
      font-size: 1rem !important;
      line-height: 1.6 !important;
    }
    
    .swagger-ui .scheme-container {
      background-color: #0f172a !important;
      border: 1px solid rgba(99, 102, 241, 0.15) !important;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2) !important;
      border-radius: 0.75rem !important;
      margin: 20px 0 !important;
      padding: 20px !important;
    }

    .swagger-ui select {
      background-color: #1e293b !important;
      color: #ffffff !important;
      border: 1px solid #475569 !important;
      border-radius: 0.375rem !important;
    }
    
    .swagger-ui .opblock {
      border-radius: 0.75rem !important;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06) !important;
      border: none !important;
      margin-bottom: 12px !important;
      overflow: hidden !important;
    }
    
    /* POST block styling */
    .swagger-ui .opblock.opblock-post {
      background: rgba(168, 85, 247, 0.05) !important;
      border-left: 5px solid #a855f7 !important;
    }
    .swagger-ui .opblock.opblock-post .opblock-summary {
      border-color: rgba(168, 85, 247, 0.15) !important;
    }
    .swagger-ui .opblock.opblock-post .opblock-summary-method {
      background-color: #a855f7 !important;
      border-radius: 0.375rem !important;
    }
    
    /* GET block styling */
    .swagger-ui .opblock.opblock-get {
      background: rgba(59, 130, 246, 0.05) !important;
      border-left: 5px solid #3b82f6 !important;
    }
    .swagger-ui .opblock.opblock-get .opblock-summary {
      border-color: rgba(59, 130, 246, 0.15) !important;
    }
    .swagger-ui .opblock.opblock-get .opblock-summary-method {
      background-color: #3b82f6 !important;
      border-radius: 0.375rem !important;
    }

    .swagger-ui .opblock .opblock-summary-path {
      color: #f8fafc !important;
      font-weight: 600 !important;
      font-size: 1.05rem !important;
    }
    
    .swagger-ui .opblock .opblock-summary-description {
      color: #94a3b8 !important;
    }
    
    .swagger-ui .btn.authorize {
      border-color: #10b981 !important;
      color: #10b981 !important;
      background: transparent !important;
      border-radius: 0.5rem !important;
      font-weight: 600 !important;
      transition: all 0.2s ease;
    }
    
    .swagger-ui .btn.authorize:hover {
      background: #10b981 !important;
      color: #ffffff !important;
    }

    .swagger-ui .btn.execute {
      background-color: #6366f1 !important;
      border-color: #6366f1 !important;
      color: white !important;
      border-radius: 0.5rem !important;
      font-weight: 600 !important;
      transition: all 0.2s ease;
    }

    .swagger-ui .btn.execute:hover {
      background-color: #4f46e5 !important;
      border-color: #4f46e5 !important;
    }
    
    .swagger-ui .topbar {
      display: none !important; /* Hide original swagger header */
    }
  </style>
</head>
<body>
  <!-- Custom Premium Header -->
  <header class="aria-header">
    <div class="aria-logo-container">
      <div class="aria-logo-icon">A</div>
      <div>
        <h1 class="aria-title">ARIA Knowledge</h1>
        <p class="aria-subtitle">Interactive API Reference</p>
      </div>
    </div>
    <div class="status-badge">
      <span class="status-dot"></span>
      API Operational
    </div>
  </header>

  <div id="swagger-ui"></div>

  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      const ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        docExpansion: "list"
      });
      window.ui = ui;
    };
  </script>
</body>
</html>
`
