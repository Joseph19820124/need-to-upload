#!/usr/bin/env node

const https = require('https');
const http = require('http');

class MCPTestClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
    this.sessionId = null;
    this.testResults = [];
  }

  log(test, status, message, data = null) {
    const result = {
      test,
      status,
      message,
      data,
      timestamp: new Date().toISOString()
    };
    this.testResults.push(result);
    
    const emoji = status === 'PASS' ? '✅' : status === 'FAIL' ? '❌' : '⏳';
    console.log(`${emoji} ${test}: ${message}`);
    if (data && status !== 'PASS') {
      console.log(`   Data: ${JSON.stringify(data, null, 2)}`);
    }
  }

  async request(path, options = {}) {
    return new Promise((resolve, reject) => {
      const fullUrl = this.baseUrl + path;
      const urlObj = new URL(fullUrl);
      const client = urlObj.protocol === 'https:' ? https : http;
      
      const reqOptions = {
        method: options.method || 'GET',
        headers: {
          'Content-Type': 'application/json',
          'User-Agent': 'MCP-Test-Client/1.0',
          ...options.headers
        }
      };

      if (this.sessionId) {
        reqOptions.headers['X-Session-ID'] = this.sessionId;
      }

      const req = client.request(fullUrl, reqOptions, (res) => {
        let data = '';
        res.on('data', chunk => data += chunk);
        res.on('end', () => {
          try {
            const result = {
              statusCode: res.statusCode,
              headers: res.headers,
              body: data,
              json: null
            };
            
            if (res.headers['content-type']?.includes('application/json')) {
              try {
                result.json = JSON.parse(data);
              } catch (e) {
                // Not valid JSON, keep as string
              }
            }
            
            resolve(result);
          } catch (e) {
            reject(e);
          }
        });
      });
      
      req.on('error', reject);
      
      if (options.body) {
        req.write(options.body);
      }
      
      req.end();
    });
  }

  async testHealth() {
    try {
      const response = await this.request('/api/v1/health');
      
      if (response.statusCode === 200) {
        this.log('Health Check', 'PASS', 'Health endpoint accessible', response.json);
        return true;
      } else {
        this.log('Health Check', 'FAIL', `HTTP ${response.statusCode}`, response.body);
        return false;
      }
    } catch (error) {
      this.log('Health Check', 'FAIL', error.message);
      return false;
    }
  }

  async testConnect() {
    try {
      const payload = {
        clientInfo: {
          name: 'mcp-test-client',
          version: '1.0.0'
        }
      };

      const response = await this.request('/api/v1/connect', {
        method: 'POST',
        body: JSON.stringify(payload)
      });

      if (response.statusCode === 200 && response.json) {
        this.sessionId = response.json.sessionId;
        this.log('Connect', 'PASS', 'Connected successfully', { 
          sessionId: this.sessionId,
          response: response.json 
        });
        return true;
      } else {
        this.log('Connect', 'FAIL', `HTTP ${response.statusCode}`, response.body);
        return false;
      }
    } catch (error) {
      this.log('Connect', 'FAIL', error.message);
      return false;
    }
  }

  async testRPC(method, params = {}) {
    try {
      const payload = {
        jsonrpc: '2.0',
        id: Date.now(),
        method: method,
        params: params
      };

      const response = await this.request('/api/v1/rpc', {
        method: 'POST',
        body: JSON.stringify(payload)
      });

      if (response.statusCode === 200 && response.json) {
        this.log(`RPC: ${method}`, 'PASS', 'RPC call successful', response.json);
        return response.json;
      } else {
        this.log(`RPC: ${method}`, 'FAIL', `HTTP ${response.statusCode}`, response.body);
        return null;
      }
    } catch (error) {
      this.log(`RPC: ${method}`, 'FAIL', error.message);
      return null;
    }
  }

  async testSSE() {
    return new Promise((resolve) => {
      try {
        const urlObj = new URL(this.baseUrl + '/api/v1/events');
        const client = urlObj.protocol === 'https:' ? https : http;
        
        const req = client.request(urlObj, {
          method: 'GET',
          headers: {
            'Accept': 'text/event-stream',
            'Cache-Control': 'no-cache',
            'X-Session-ID': this.sessionId || 'test-session'
          }
        }, (res) => {
          if (res.statusCode === 200) {
            this.log('SSE Connection', 'PASS', 'SSE stream connected');
            
            let dataReceived = false;
            const timeout = setTimeout(() => {
              if (!dataReceived) {
                this.log('SSE Data', 'INFO', 'No SSE data received within 5 seconds (this is normal)');
              }
              req.destroy();
              resolve(true);
            }, 5000);

            res.on('data', (chunk) => {
              dataReceived = true;
              const data = chunk.toString();
              this.log('SSE Data', 'PASS', 'Received SSE data', { data: data.trim() });
              clearTimeout(timeout);
              req.destroy();
              resolve(true);
            });

            res.on('end', () => {
              clearTimeout(timeout);
              resolve(true);
            });
          } else {
            this.log('SSE Connection', 'FAIL', `HTTP ${res.statusCode}`);
            resolve(false);
          }
        });

        req.on('error', (error) => {
          this.log('SSE Connection', 'FAIL', error.message);
          resolve(false);
        });

        req.end();
      } catch (error) {
        this.log('SSE Connection', 'FAIL', error.message);
        resolve(false);
      }
    });
  }

  async runAllTests() {
    console.log('🚀 Starting MCP Server Tests');
    console.log(`📡 Testing server: ${this.baseUrl}`);
    console.log('=' .repeat(60));

    // Test 1: Health check
    await this.testHealth();

    // Test 2: Connect
    const connected = await this.testConnect();

    // Test 3: SSE connection
    await this.testSSE();

    // Test 4: Basic RPC calls (only if connected)
    if (connected) {
      await this.testRPC('ping');
      await this.testRPC('tools/list');
      await this.testRPC('resources/list');
    }

    // Test 5: Disconnect
    if (this.sessionId) {
      await this.request('/api/v1/disconnect', {
        method: 'POST'
      });
      this.log('Disconnect', 'INFO', 'Disconnect request sent');
    }

    // Summary
    console.log('=' .repeat(60));
    console.log('📊 Test Summary:');
    
    const passed = this.testResults.filter(r => r.status === 'PASS').length;
    const failed = this.testResults.filter(r => r.status === 'FAIL').length;
    const info = this.testResults.filter(r => r.status === 'INFO').length;
    
    console.log(`✅ Passed: ${passed}`);
    console.log(`❌ Failed: ${failed}`);
    console.log(`ℹ️  Info: ${info}`);
    
    if (failed === 0) {
      console.log('🎉 All tests passed! Your MCP server is working correctly.');
    } else {
      console.log('⚠️  Some tests failed. Check the details above.');
    }

    return this.testResults;
  }
}

// Main execution
async function main() {
  const serverUrl = process.argv[2] || 'https://calm-benevolence-production.up.railway.app';
  
  console.log('🧪 MCP Server Test Client');
  console.log('Usage: node test-mcp-client.js [server-url]');
  console.log('');

  const client = new MCPTestClient(serverUrl);
  await client.runAllTests();
}

if (require.main === module) {
  main().catch(console.error);
}

module.exports = MCPTestClient;