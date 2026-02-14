#!/usr/bin/env node

/**
 * WebSocket Chat Test Script for Order #1
 * Tests both Customer and Mitra connections
 */

const WebSocket = require('ws');
const readline = require('readline');

// Tokens
const CUSTOMER_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InJhbnRpaXNzYXJhNzJAZ21haWwuY29tIiwiZXhwIjoxNzcxMTYzNDczLCJpZCI6NywibmFtYSI6IlJhbnRpIElzc2FyYSIsInBob25lIjoiMDgxMzk4ODMyODMxIn0.nGTEDnlz5ztpRRPtnpqe0CmeGtq5vTf-LGpwTlQ-toY';
const MITRA_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFuZHJpcHJhc3V0aW9AZ21haWwuY29tIiwiZXhwIjoxNzcxMTYzNDMwLCJpZCI6NiwibmFtYSI6IkFuZHJpIFByYXN1dGlvIiwicGhvbmUiOiIwODEzOTg4MzI4MzAifQ.FoYn6r-z0MNBSJX3298mlrjbDAlA_Z9ixYXUZ_-Wocw';

// Configuration
const SERVER_URL = 'ws://localhost:8080/api/realtime/chat/2';
// const SERVER_URL = 'wss://be-teka-katanyangoding255248-afaak30s.leapcell.dev/api/realtime/chat/2';
const ORDER_ID = 2;

class ChatClient {
    constructor(name, token, color) {
        this.name = name;
        this.token = token;
        this.color = color;
        this.ws = null;
        this.connected = false;
        this.historyCount = 0;
        this.isLoadingHistory = true;
    }

    connect() {
        return new Promise((resolve, reject) => {
            const wsUrl = `${SERVER_URL}?token=${this.token}`;
            console.log(`${this.color}🔌 [${this.name}] Connecting to ${SERVER_URL}...${'\x1b[0m'}`);
            
            this.ws = new WebSocket(wsUrl);

            this.ws.on('open', () => {
                this.connected = true;
                console.log(`${this.color}✅ [${this.name}] Connected successfully${'\x1b[0m'}`);
                resolve();
            });

            this.ws.on('message', (data) => {
                try {
                    const message = JSON.parse(data.toString());
                    this.handleMessage(message);
                } catch (e) {
                    console.error(`${this.color}❌ [${this.name}] Error parsing message:${'\x1b[0m'}`, e.message);
                }
            });

            this.ws.on('error', (error) => {
                console.error(`${this.color}❌ [${this.name}] WebSocket error:${'\x1b[0m'}`, error.message);
                reject(error);
            });

            this.ws.on('close', (code, reason) => {
                this.connected = false;
                console.log(`${this.color}🔌 [${this.name}] Disconnected (code: ${code}, reason: ${reason || 'none'})${'\x1b[0m'}`);
            });
        });
    }

    handleMessage(message) {
        if (message.error) {
            console.error(`${this.color}❌ [${this.name}] Error: ${message.error}${'\x1b[0m'}`);
            return;
        }

        const time = new Date(message.created_at).toLocaleTimeString();
        
        // Track history messages (messages received right after connection)
        if (this.isLoadingHistory) {
            this.historyCount++;
            const prefix = message.sender_type === this.name.toLowerCase() ? '📤 SENT' : '📥 RECEIVED';
            console.log(`${this.color}📜 HISTORY [${this.name}] ${message.sender_type} at ${time}: ${message.message}${'\x1b[0m'}`);
            
            // After a short delay, assume history loading is done
            setTimeout(() => {
                if (this.isLoadingHistory && this.historyCount > 0) {
                    this.isLoadingHistory = false;
                    console.log(`${this.color}✅ [${this.name}] Loaded ${this.historyCount} messages from history${'\x1b[0m'}`);
                }
            }, 500);
        } else {
            // Real-time messages
            const prefix = message.sender_type === this.name.toLowerCase() ? '📤 SENT' : '📥 RECEIVED';
            console.log(`${this.color}${prefix} [${this.name}] ${message.sender_type} at ${time}: ${message.message}${'\x1b[0m'}`);
        }
    }

    sendMessage(text) {
        if (!this.connected || !this.ws) {
            console.error(`${this.color}❌ [${this.name}] Not connected${'\x1b[0m'}`);
            return false;
        }

        const payload = { message: text };
        this.ws.send(JSON.stringify(payload));
        console.log(`${this.color}📤 [${this.name}] Sent: ${text}${'\x1b[0m'}`);
        return true;
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Initialize clients
const customerClient = new ChatClient('Customer', CUSTOMER_TOKEN, '\x1b[36m'); // Cyan
const mitraClient = new ChatClient('Mitra', MITRA_TOKEN, '\x1b[35m'); // Magenta

// Interactive CLI
const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
});

function showMenu() {
    console.log('\n\x1b[33m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m');
    console.log('\x1b[1m📱 WebSocket Chat Test Menu\x1b[0m');
    console.log('\x1b[33m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m');
    console.log('1. Connect Customer');
    console.log('2. Connect Mitra');
    console.log('3. Send message as Customer');
    console.log('4. Send message as Mitra');
    console.log('5. Disconnect Customer');
    console.log('6. Disconnect Mitra');
    console.log('7. Auto test (send messages from both)');
    console.log('8. Exit');
    console.log('\x1b[33m━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\x1b[0m\n');
}

async function autoTest() {
    console.log('\n\x1b[32m🤖 Starting automated test...\x1b[0m\n');
    
    try {
        // Connect both
        await customerClient.connect();
        await new Promise(resolve => setTimeout(resolve, 500));
        await mitraClient.connect();
        await new Promise(resolve => setTimeout(resolve, 1000));

        // Send test messages
        const messages = [
            { sender: customerClient, text: 'Hello! I need help with my order.' },
            { sender: mitraClient, text: 'Hi! I\'m on my way to your location.' },
            { sender: customerClient, text: 'Great! How long will it take?' },
            { sender: mitraClient, text: 'About 10 minutes. I\'ll be there soon!' },
            { sender: customerClient, text: 'Perfect, thank you!' }
        ];

        for (const msg of messages) {
            await new Promise(resolve => setTimeout(resolve, 1500));
            msg.sender.sendMessage(msg.text);
        }

        console.log('\n\x1b[32m✅ Automated test completed!\x1b[0m\n');
    } catch (error) {
        console.error('\n\x1b[31m❌ Automated test failed:\x1b[0m', error.message);
    }
}

function promptMessage(client) {
    rl.question(`Enter message for ${client.name}: `, (text) => {
        if (text.trim()) {
            client.sendMessage(text.trim());
        }
        handleMenu();
    });
}

function handleMenu() {
    showMenu();
    rl.question('Select option: ', async (answer) => {
        switch (answer.trim()) {
            case '1':
                await customerClient.connect().catch(e => console.error('Connection failed:', e.message));
                handleMenu();
                break;
            case '2':
                await mitraClient.connect().catch(e => console.error('Connection failed:', e.message));
                handleMenu();
                break;
            case '3':
                promptMessage(customerClient);
                break;
            case '4':
                promptMessage(mitraClient);
                break;
            case '5':
                customerClient.disconnect();
                handleMenu();
                break;
            case '6':
                mitraClient.disconnect();
                handleMenu();
                break;
            case '7':
                await autoTest();
                handleMenu();
                break;
            case '8':
                console.log('\n\x1b[32m👋 Goodbye!\x1b[0m\n');
                customerClient.disconnect();
                mitraClient.disconnect();
                rl.close();
                process.exit(0);
                break;
            default:
                console.log('\x1b[31m❌ Invalid option\x1b[0m');
                handleMenu();
        }
    });
}

// Handle cleanup
process.on('SIGINT', () => {
    console.log('\n\n\x1b[33m⚠️  Shutting down...\x1b[0m');
    customerClient.disconnect();
    mitraClient.disconnect();
    rl.close();
    process.exit(0);
});

// Start
console.log('\x1b[1m\n🚀 WebSocket Chat Test for Order #2\x1b[0m');
console.log(`\x1b[90mServer: ${SERVER_URL}\x1b[0m`);
console.log(`\x1b[90mCustomer: Ranti Issara (ID: 7)\x1b[0m`);
console.log(`\x1b[90mMitra: Andri Prasutio (ID: 6)\x1b[0m\n`);

handleMenu();
