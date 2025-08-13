import React, { useState, useEffect, useRef } from "react";
import { PaperPlaneIcon } from "@radix-ui/react-icons";

const BASE_URL = "http://localhost:8080";
const USER_ID = "12345678-0000-0000-0000-000000000000";

export default function RAGChatUI() {
    const [messages, setMessages] = useState([]);
    const [input, setInput] = useState("");
    const [loading, setLoading] = useState(false);
    const [token, setToken] = useState("");
    const [chatId, setChatId] = useState("");
    const messagesEndRef = useRef(null);

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages]);

    useEffect(() => {
        const authenticateAndStartChat = async () => {
            try {
                const tokenResponse = await fetch(`${BASE_URL}/token`, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ username: "user", password: "password" }),
                });

                const tokenData = await tokenResponse.json();
                if (!tokenData.token) throw new Error("Token retrieval failed.");
                setToken(tokenData.token);

                const chatResponse = await fetch(`${BASE_URL}/users/${USER_ID}/chat-sessions`, {
                    method: "POST",
                    headers: { Authorization: `Bearer ${tokenData.token}` },
                });

                const chatData = await chatResponse.json();
                if (!chatData.id) throw new Error("Chat session creation failed.");
                setChatId(chatData.id);
            } catch (err) {
                console.error(err);
            }
        };

        authenticateAndStartChat();
    }, []);

    const sendMessage = async () => {
        if (!input.trim() || !token || !chatId) return;

        const userMessage = { role: "user", content: input };
        setMessages([userMessage, { role: "assistant", content: "" }]);
        const prompt = input;
        setInput("");
        setLoading(true);

        try {
            const response = await fetch(
                `${BASE_URL}/sse/users/${USER_ID}/chat-sessions/${chatId}/messages`,
                {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                        Authorization: `Bearer ${token}`,
                    },
                    body: JSON.stringify({ content: prompt }),
                }
            );

            if (!response.ok || !response.body) {
                throw new Error("Streaming failed");
            }

            const reader = response.body.getReader();
            const decoder = new TextDecoder("utf-8");
            let fullMessage = "";
            let buffer = "";

            while (true) {
                const { value, done } = await reader.read();
                if (done) break;
                buffer += decoder.decode(value, { stream: true });

                const lines = buffer.split(/\n/);
                buffer = lines.pop() || "";

                for (const line of lines) {
                    if (line.startsWith("data:")) {
                        const json = line.slice(5).trim();
                        try {
                            const parsed = JSON.parse(json);
                            if (parsed.token) {
                                fullMessage += parsed.token + " ";
                                let extracted = "";
                                const contentMatch = fullMessage.match(/Content:(.*?)CreatedAt:/);
                                if (contentMatch) {
                                    extracted = contentMatch[1].trim();
                                } else if (fullMessage.includes("Content:")) {
                                    extracted = fullMessage.split("Content:").pop().trim();
                                } else {
                                    extracted = "";
                                }
                                setMessages((prev) => [prev[0], { role: "assistant", content: extracted }]);
                            }
                        } catch (e) {
                            console.warn("Skipping invalid JSON:", json);
                        }
                    }
                }
            }
        } catch (err) {
            console.error("Streaming message failed:", err);
            setMessages([messages[0], { role: "assistant", content: "Error retrieving response." }]);
        }

        setLoading(false);
    };

    const handleKeyDown = (e) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            sendMessage();
        }
    };

    return (
        <div style={{
            minHeight: "100vh",
            display: "flex",
            flexDirection: "column",
            justifyContent: "center",
            alignItems: "center",
            background: "linear-gradient(to bottom right, #c7d2fe, #e0f2fe, #ffffff)",
            fontFamily: "Arial, sans-serif",
            padding: "2rem"
        }}>
            <h1 style={{
                fontSize: "3.5rem",
                fontWeight: "800",
                textAlign: "center",
                color: "#1e40af",
                marginBottom: "2rem",
                letterSpacing: "0.05em"
            }}>
                Louk Chatwalker
            </h1>
            <div style={{
                width: "100%",
                maxWidth: "700px",
                background: "#ffffff",
                borderRadius: "1.5rem",
                boxShadow: "0 20px 40px rgba(0, 0, 0, 0.1)",
                padding: "2rem",
                textAlign: "center"
            }}>
        <textarea
            style={{
                width: "100%",
                textAlign: "center",
                fontSize: "1.5rem",
                padding: "1.25rem",
                borderRadius: "1rem",
                border: "1px solid #d1d5db",
                resize: "none",
                marginBottom: "1.5rem",
                outline: "none"
            }}
            rows="2"
            placeholder="Ask your question..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
        ></textarea>
                <button
                    style={{
                        marginBottom: "1.5rem",
                        padding: "1rem 2rem",
                        background: "linear-gradient(to right, #0ea5e9, #2563eb)",
                        color: "white",
                        fontSize: "1.1rem",
                        fontWeight: "500",
                        borderRadius: "1rem",
                        border: "none",
                        cursor: "pointer",
                        opacity: loading ? 0.6 : 1
                    }}
                    onClick={sendMessage}
                    disabled={loading || !token || !chatId}
                >
                    {loading ? (
                        <svg style={{ height: "1.25rem", width: "1.25rem" }} className="animate-spin" viewBox="0 0 24 24">
                            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
                        </svg>
                    ) : (
                        <PaperPlaneIcon style={{ height: "1.25rem", width: "1.25rem" }} />
                    )}
                </button>
                <div style={{
                    width: "100%",
                    padding: "1.5rem",
                    background: "#f3f4f6",
                    borderRadius: "1rem",
                    minHeight: "150px",
                    textAlign: "left",
                    fontSize: "1.1rem",
                    lineHeight: "1.6",
                    fontWeight: "300",
                    whiteSpace: "pre-wrap"
                }}>
                    {messages[1]?.content || "The answer will appear here..."}
                </div>
            </div>
        </div>
    );
}
