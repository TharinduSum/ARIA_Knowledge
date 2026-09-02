document.addEventListener("DOMContentLoaded", () => {
  const chatWindow = document.getElementById("chat-window");
  const userInput = document.getElementById("user-input");
  const sendBtn = document.getElementById("send-btn");
  const streamToggle = document.getElementById("stream-toggle");
  const topKSelect = document.getElementById("top-k-select");
  const dbStatusDot = document.getElementById("db-status-dot");
  const dbStatusText = document.getElementById("db-status-text");
  const themeToggleBtn = document.getElementById("theme-toggle-btn");

  let chatHistory = [];
  let isGenerating = false;

  // Configure marked parser if loaded
  if (window.marked) {
    const renderer = new marked.Renderer();

    renderer.link = function (href, title, text) {
      if (typeof href === "object" && href !== null) {
        // Marked v12 token structure
        const token = href;
        return `<a href="${token.href}" target="_blank" rel="noopener noreferrer"${token.title ? ` title="${token.title}"` : ""}>${token.text}</a>`;
      }
      return `<a href="${href}" target="_blank" rel="noopener noreferrer"${title ? ` title="${title}"` : ""}>${text}</a>`;
    };

    marked.setOptions({
      renderer: renderer,
      breaks: true,
      gfm: true,
      highlight: function (code, lang) {
        if (window.hljs && lang && window.hljs.getLanguage(lang)) {
          try {
            return window.hljs.highlight(code, { language: lang }).value;
          } catch (e) {}
        }
        if (window.hljs) {
          try {
            return window.hljs.highlightAuto(code).value;
          } catch (e) {}
        }
        return code;
      },
    });
  }

  // Markdown Converter Helper
  function renderMarkdown(text) {
    if (!text) return "";
    try {
      let rawHtml = "";
      if (window.marked && typeof window.marked.parse === "function") {
        rawHtml = window.marked.parse(text);
      } else {
        rawHtml = fallbackMarkdown(text);
      }

      if (window.DOMPurify && typeof window.DOMPurify.sanitize === "function") {
        return window.DOMPurify.sanitize(rawHtml, {
          ADD_ATTR: ["target", "rel"],
        });
      }
      return rawHtml;
    } catch (err) {
      console.error("Markdown parsing error:", err);
      return escapeHtml(text);
    }
  }

  function fallbackMarkdown(text) {
    let html = escapeHtml(text);
    html = html.replace(/\n\n+/g, "</p><p>");
    html = html.replace(/\*\*(.*?)\*\*/g, "<strong>$1</strong>");
    html = html.replace(/\*(.*?)\*/g, "<em>$1</em>");
    html = html.replace(/`([^`]+)`/g, "<code>$1</code>");
    return `<p>${html}</p>`;
  }

  // Enhance Code Blocks with Copy button and Language Header
  function enhanceCodeBlocks(container) {
    if (!container) return;
    const preElements = container.querySelectorAll("pre");

    preElements.forEach((pre) => {
      if (pre.parentElement && pre.parentElement.classList.contains("code-block-wrapper")) {
        return;
      }

      const code = pre.querySelector("code");
      const codeText = code ? code.innerText : pre.innerText;

      let lang = "code";
      if (code && code.className) {
        const match = code.className.match(/language-([a-zA-Z0-9_\-]+)/);
        if (match) lang = match[1];
      }

      const wrapper = document.createElement("div");
      wrapper.className = "code-block-wrapper";

      const header = document.createElement("div");
      header.className = "code-block-header";

      const langSpan = document.createElement("span");
      langSpan.textContent = lang;

      const copyBtn = document.createElement("button");
      copyBtn.className = "code-copy-btn";
      copyBtn.type = "button";
      copyBtn.innerHTML = `
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
        </svg>
        <span>Copy</span>
      `;

      copyBtn.addEventListener("click", async () => {
        try {
          await navigator.clipboard.writeText(codeText);
          copyBtn.innerHTML = `
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#10b981" stroke-width="2">
              <polyline points="20 6 9 17 4 12"></polyline>
            </svg>
            <span style="color: #10b981; font-weight: 500;">Copied!</span>
          `;
          setTimeout(() => {
            copyBtn.innerHTML = `
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
                <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
              </svg>
              <span>Copy</span>
            `;
          }, 2000);
        } catch (err) {
          console.error("Failed to copy code snippet: ", err);
        }
      });

      header.appendChild(langSpan);
      header.appendChild(copyBtn);

      pre.parentNode.insertBefore(wrapper, pre);
      wrapper.appendChild(header);
      wrapper.appendChild(pre);
    });
  }

  // Theme Management (Light Theme Default)
  const savedTheme = localStorage.getItem("aria_theme") || "light";
  document.documentElement.setAttribute("data-theme", savedTheme === "dark" ? "dark" : "light");

  if (themeToggleBtn) {
    themeToggleBtn.addEventListener("click", () => {
      const currentTheme = document.documentElement.getAttribute("data-theme");
      const newTheme = currentTheme === "dark" ? "light" : "dark";
      document.documentElement.setAttribute("data-theme", newTheme);
      localStorage.setItem("aria_theme", newTheme);
    });
  }

  // Auto resize textarea
  userInput.addEventListener("input", () => {
    userInput.style.height = "auto";
    userInput.style.height = Math.min(userInput.scrollHeight, 120) + "px";
  });

  // Handle Enter key (Shift+Enter for newline)
  userInput.addEventListener("keydown", (e) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  });

  sendBtn.addEventListener("click", sendMessage);

  // Fetch Health Status
  async function checkHealth() {
    try {
      const res = await fetch("/health");
      const data = await res.json();
      if (data.status === "healthy") {
        dbStatusDot.classList.remove("offline");
        dbStatusText.textContent = `Online (${data.ollama})`;
      } else {
        dbStatusDot.classList.add("offline");
        dbStatusText.textContent = `Degraded (${data.ollama})`;
      }
    } catch (err) {
      dbStatusDot.classList.add("offline");
      dbStatusText.textContent = "Offline";
    }
  }

  checkHealth();
  setInterval(checkHealth, 30000);

  function appendMessage(role, text, citations = []) {
    const row = document.createElement("div");
    row.className = `message-row ${role}`;

    const avatar = document.createElement("div");
    avatar.className = "avatar";
    avatar.textContent = role === "user" ? "U" : "AI";

    const bubble = document.createElement("div");
    bubble.className = "message-bubble";

    if (role === "assistant") {
      bubble.classList.add("markdown-body");
      bubble.innerHTML = renderMarkdown(text);
      enhanceCodeBlocks(bubble);
    } else {
      bubble.textContent = text;
    }

    if (citations && citations.length > 0) {
      appendCitationsToBubble(bubble, citations);
    }

    row.appendChild(avatar);
    row.appendChild(bubble);
    chatWindow.appendChild(row);
    scrollToBottom();
    return bubble;
  }

  function escapeHtml(text) {
    const div = document.createElement("div");
    div.textContent = text;
    return div.innerHTML;
  }

  function scrollToBottom() {
    chatWindow.scrollTop = chatWindow.scrollHeight;
  }

  async function sendMessage() {
    const message = userInput.value.trim();
    if (!message || isGenerating) return;

    // Reset input
    userInput.value = "";
    userInput.style.height = "24px";
    isGenerating = true;
    sendBtn.disabled = true;

    // Append User Message
    appendMessage("user", message);
    chatHistory.push({ role: "user", content: message });

    const isStreaming = streamToggle.checked;
    const topK = parseInt(topKSelect.value, 10);

    if (isStreaming) {
      await handleStreamingChat(message, topK);
    } else {
      await handleStandardChat(message, topK);
    }

    isGenerating = false;
    sendBtn.disabled = false;
    userInput.focus();
  }

  // Standard JSON Chat Request
  async function handleStandardChat(message, topK) {
    const assistantBubble = appendMessage("assistant", "Thinking...");

    try {
      const response = await fetch("/api/v1/chat", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          message: message,
          history: chatHistory.slice(-6),
          top_k: topK,
          stream: false,
        }),
      });

      if (!response.ok) {
        throw new Error(`Server returned status ${response.status}`);
      }

      const data = await response.json();
      assistantBubble.innerHTML = renderMarkdown(data.message || data.html_message);
      enhanceCodeBlocks(assistantBubble);

      if (data.citations && data.citations.length > 0) {
        appendCitationsToBubble(assistantBubble, data.citations);
      }

      chatHistory.push({ role: "assistant", content: data.message });
      scrollToBottom();
    } catch (err) {
      assistantBubble.innerHTML = `<span style="color: var(--danger-color);">Error: ${escapeHtml(err.message)}</span>`;
    }
  }

  // Streaming SSE Chat Request
  async function handleStreamingChat(message, topK) {
    const row = document.createElement("div");
    row.className = "message-row assistant";

    const avatar = document.createElement("div");
    avatar.className = "avatar";
    avatar.textContent = "AI";

    const bubble = document.createElement("div");
    bubble.className = "message-bubble markdown-body";

    const contentContainer = document.createElement("div");
    contentContainer.className = "streaming-content";

    const cursorSpan = document.createElement("span");
    cursorSpan.className = "cursor";

    bubble.appendChild(contentContainer);
    bubble.appendChild(cursorSpan);

    row.appendChild(avatar);
    row.appendChild(bubble);
    chatWindow.appendChild(row);
    scrollToBottom();

    let fullAnswer = "";
    let citations = [];

    try {
      const response = await fetch("/api/v1/chat/stream", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          message: message,
          history: chatHistory.slice(-6),
          top_k: topK,
          stream: true,
        }),
      });

      if (!response.ok) {
        throw new Error(`Streaming failed: ${response.status}`);
      }

      const reader = response.body.getReader();
      const decoder = new TextDecoder();
      let buffer = "";

      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n\n");
        buffer = lines.pop(); // Keep incomplete line in buffer

        for (const line of lines) {
          if (line.startsWith("data: ")) {
            const jsonStr = line.replace("data: ", "").trim();
            if (!jsonStr) continue;

            try {
              const eventData = JSON.parse(jsonStr);

              // Handle citations metadata
              if (eventData.type === "citations") {
                citations = eventData.citations;
              }
              // Handle text token
              else if (eventData.token) {
                fullAnswer += eventData.token;
                contentContainer.innerHTML = renderMarkdown(fullAnswer);
                scrollToBottom();
              }
            } catch (e) {
              console.error("Failed to parse SSE payload", e);
            }
          }
        }
      }

      // Remove blinking cursor when done
      cursorSpan.remove();

      // Final complete render & enhance code blocks
      contentContainer.innerHTML = renderMarkdown(fullAnswer);
      enhanceCodeBlocks(contentContainer);

      if (citations && citations.length > 0) {
        appendCitationsToBubble(bubble, citations);
      }

      chatHistory.push({ role: "assistant", content: fullAnswer });
      scrollToBottom();

    } catch (err) {
      cursorSpan.remove();
      contentContainer.innerHTML = `<span style="color: var(--danger-color);">Streaming Error: ${escapeHtml(err.message)}</span>`;
    }
  }

  function appendCitationsToBubble(bubble, citations) {
    const citationsContainer = document.createElement("div");
    citationsContainer.className = "citations-container";

    const title = document.createElement("div");
    title.className = "citations-title";
    title.innerHTML = `📚 <span>Sources (${citations.length})</span>`;
    citationsContainer.appendChild(title);

    citations.forEach((c) => {
      const card = document.createElement("div");
      card.className = "citation-card";

      const sourceName = c.metadata?.source || c.metadata?.filename || "Document";
      const pageNum = c.metadata?.page || c.metadata?.page_number;
      const pageStr = pageNum ? ` • Page ${pageNum}` : "";
      const simScore = (c.similarity * 100).toFixed(1);

      card.innerHTML = `
        <div class="citation-meta">📄 ${sourceName}${pageStr} (${simScore}% match)</div>
        <div>${escapeHtml(c.content.slice(0, 150))}...</div>
      `;
      citationsContainer.appendChild(card);
    });

    bubble.appendChild(citationsContainer);
  }
});
