// React 19 requires MessageChannel which jsdom doesn't expose globally
const { MessageChannel } = require('worker_threads');
global.MessageChannel = MessageChannel;

// Jest setup provided by Grafana scaffolding
import './.config/jest-setup';
