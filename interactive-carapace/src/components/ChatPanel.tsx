'use client';

import React, { useState, useEffect, useRef } from 'react';
import { Send, Bot, User, Loader2, MessageSquare, AlertCircle, Wifi, WifiOff, WifiLow, RotateCw } from 'lucide-react';
import { useWebSocket } from '@/hooks/useWebSocket';
import { useMessages, useConnectionState, useSessionId, useError, useIsTyping } from '@/store/chatStore';

export function ChatPanel() {
    const [inputMessage, setInputMessage] = useState('');
    const [isInputFocused, setIsInputFocused] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    // WebSocket hook
    const { sendMessage, isConnected, connectionState, error: wsError } = useWebSocket();

    // Zustand store selectors
    const messages = useMessages();
    const sessionId = useSessionId();
    const storeError = useError();
    const isTyping = useIsTyping();

    // Use the error from either WebSocket hook or store
    const currentError = wsError || storeError;

    // Auto-scroll to bottom when new messages arrive
    useEffect(() => {
        scrollToBottom();
    }, [messages, isTyping]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    const handleSendMessage = () => {
        if (!inputMessage.trim() || !isConnected) return;

        sendMessage({
            type: 'chat',
            content: inputMessage.trim(),
            author: 'user',
        });

        setInputMessage('');
    };

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSendMessage();
        }
    };

    const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setInputMessage(e.target.value);
        
        // Auto-resize textarea
        e.target.style.height = 'auto';
        e.target.style.height = Math.min(e.target.scrollHeight, 120) + 'px';
    };

    const getConnectionStatusIcon = () => {
        switch (connectionState) {
            case 'connected': return <Wifi className="w-4 h-4 text-green-500" />;
            case 'connecting': return <RotateCw className="w-4 h-4 text-yellow-500 animate-spin" />;
            case 'error': return <WifiOff className="w-4 h-4 text-red-500" />;
            default: return <WifiLow className="w-4 h-4 text-gray-400" />;
        }
    };

    const getConnectionStatusColor = () => {
        switch (connectionState) {
            case 'connected': return 'bg-green-400';
            case 'connecting': return 'bg-yellow-400';
            case 'error': return 'bg-red-400';
            default: return 'bg-gray-400';
        }
    };

    const getConnectionStatusText = () => {
        switch (connectionState) {
            case 'connected': return 'Connected';
            case 'connecting': return 'Connecting...';
            case 'error': return 'Connection Error';
            default: return 'Disconnected';
        }
    };

    // Focus textarea when component mounts
    useEffect(() => {
        if (textareaRef.current) {
            textareaRef.current.focus();
        }
    }, []);

    return (
        <div className="flex flex-col h-full bg-white">
            {/* Chat Header */}
            <div className="px-4 py-3 border-b border-gray-200">
                <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-2">
                        <MessageSquare className="w-5 h-5 text-gray-600" />
                        <h3 className="font-medium text-gray-900">AI Assistant</h3>
                    </div>
                    <div className="flex items-center space-x-2">
                        {getConnectionStatusIcon()}
                        <span className="text-xs text-gray-500">
                            {getConnectionStatusText()}
                        </span>
                    </div>
                </div>

                {/* Error Display */}
                {(currentError || connectionState === 'error') && (
                    <div className="mt-2 flex items-center space-x-2 text-xs text-red-600 bg-red-50 px-2 py-1 rounded">
                        <AlertCircle className="w-3 h-3 flex-shrink-0" />
                        <span>{currentError || 'Connection error. Please check your network and try again.'}</span>
                    </div>
                )}

                {/* Session Info */}
                {sessionId && (
                    <div className="mt-1 text-xs text-gray-400">
                        Session: {sessionId.slice(-8)}
                    </div>
                )}
            </div>

            {/* Messages */}
            <div className="flex-1 overflow-y-auto p-4 space-y-4 chat-scrollbar">
                {messages.length === 0 ? (
                    <div className="flex flex-col items-center justify-center h-full text-center text-gray-500">
                        <Bot className="w-12 h-12 mb-3 text-gray-300" />
                        <h3 className="text-lg font-medium mb-1">Welcome to Aether AI Assistant</h3>
                        <p className="text-sm max-w-xs">
                            Upload a CSV file and ask me questions about your data. I can help you create visualizations and analyze trends.
                        </p>
                    </div>
                ) : (
                    messages.map((msg) => (
                        <div
                            key={msg.id}
                            className={`flex items-start space-x-3 ${msg.author === 'user' ? 'flex-row-reverse space-x-reverse' : ''}`}
                        >
                            <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${
                                msg.author === 'user'
                                    ? 'bg-blue-100'
                                    : msg.author === 'ai'
                                        ? 'bg-purple-100'
                                        : 'bg-gray-100'
                            }`}>
                                {msg.author === 'user' ? (
                                    <User className="w-4 h-4 text-blue-600" />
                                ) : msg.author === 'ai' ? (
                                    <Bot className="w-4 h-4 text-purple-600" />
                                ) : (
                                    <MessageSquare className="w-4 h-4 text-gray-600" />
                                )}
                            </div>

                            <div className={`flex-1 max-w-xs ${msg.author === 'user' ? 'text-right' : ''}`}>
                                <div className={`inline-block px-3 py-2 rounded-lg text-sm ${
                                    msg.author === 'user'
                                        ? 'bg-blue-600 text-white'
                                        : msg.author === 'ai'
                                            ? 'bg-gray-100 text-gray-900'
                                            : 'bg-yellow-50 text-yellow-800 border border-yellow-200'
                                }`}>
                                    {msg.content}
                                </div>
                                <div className="text-xs text-gray-500 mt-1">
                                    {msg.timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                                </div>
                            </div>
                        </div>
                    ))
                )}

                {/* Typing Indicator */}
                {isTyping && (
                    <div className="flex items-start space-x-3">
                        <div className="flex-shrink-0 w-8 h-8 rounded-full bg-purple-100 flex items-center justify-center">
                            <Bot className="w-4 h-4 text-purple-600" />
                        </div>
                        <div className="flex items-center space-x-2 px-3 py-2 bg-gray-100 rounded-lg">
                            <Loader2 className="w-4 h-4 animate-spin text-gray-500" />
                            <span className="text-sm text-gray-500">AI is thinking...</span>
                        </div>
                    </div>
                )}

                {/* Auto-scroll anchor */}
                <div ref={messagesEndRef} />
            </div>

            {/* Message Input */}
            <div className="px-4 py-3 border-t border-gray-200">
                <div className="flex items-center space-x-2">
                    <textarea
                        ref={textareaRef}
                        value={inputMessage}
                        onChange={handleInputChange}
                        onKeyPress={handleKeyPress}
                        onFocus={() => setIsInputFocused(true)}
                        onBlur={() => setIsInputFocused(false)}
                        placeholder={
                            isConnected
                                ? "Ask me about your data..."
                                : "Connecting to server..."
                        }
                        disabled={!isConnected}
                        className="flex-1 resize-none border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:bg-gray-50 disabled:text-gray-400"
                        rows={1}
                        style={{ minHeight: '36px', maxHeight: '120px' }}
                    />
                    <button
                        onClick={handleSendMessage}
                        disabled={!inputMessage.trim() || !isConnected || isTyping}
                        className="p-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                        aria-label="Send message"
                    >
                        <Send className="w-4 h-4" />
                    </button>
                </div>

                {/* Connection Status Help Text */}
                {!isConnected && (
                    <div className="mt-2 text-xs text-gray-500">
                        {connectionState === 'connecting'
                            ? 'Establishing connection to AI assistant...'
                            : 'Unable to connect to AI assistant. Please check your connection.'}
                    </div>
                )}

                {/* Input hint */}
                {isInputFocused && isConnected && (
                    <div className="mt-1 text-xs text-gray-400 flex items-center">
                        <kbd className="px-1 py-0.5 text-xs rounded bg-gray-100">Enter</kbd>
                        <span className="ml-1">to send</span>
                    </div>
                )}
            </div>
        </div>
    );
}