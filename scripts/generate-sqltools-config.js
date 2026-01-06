#!/usr/bin/env node
/**
 * Converts infra-mcp-server config.json to SQLTools VS Code extension format.
 * SQLTools stores connections in settings.json - much simpler than Database Client!
 * 
 * Usage:
 *   node scripts/generate-sqltools-config.js [options] [config-file]
 * 
 * Options:
 *   --write    Write directly to .vscode/settings.json
 *   --print    Print JSON to stdout (default)
 * 
 * Examples:
 *   node scripts/generate-sqltools-config.js                     # Print to stdout
 *   node scripts/generate-sqltools-config.js --write             # Write to .vscode/settings.json
 *   node scripts/generate-sqltools-config.js bin/config.json     # Use custom config file
 */

const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
const writeMode = args.includes('--write');
const configArg = args.find(arg => !arg.startsWith('--'));
const configPath = configArg || path.join(__dirname, '../bin/config.json');
const settingsPath = path.join(__dirname, '../.vscode/settings.json');

function convertConnection(conn) {
    // SQLTools uses 'PostgreSQL' or 'MySQL' as driver names
    // Driver package names: mtxr.sqltools-driver-pg, mtxr.sqltools-driver-mysql
    const driver = conn.type === 'postgres' ? 'PostgreSQL' : 'MySQL';
    
    const connection = {
        name: conn.display_name || conn.id,
        driver: driver,
        server: conn.host,
        port: conn.port,
        database: conn.name,
        username: conn.user,
        password: conn.password,
        group: conn.environment || 'default'
    };
    
    // Add PostgreSQL-specific settings
    if (conn.type === 'postgres') {
        connection.previewLimit = 50;
        connection.connectionTimeout = 15;
    }
    
    return connection;
}

function main() {
    if (!fs.existsSync(configPath)) {
        console.error(`Config file not found: ${configPath}`);
        process.exit(1);
    }
    
    const configData = fs.readFileSync(configPath, 'utf8');
    const config = JSON.parse(configData);
    
    // Convert all connections
    const connections = (config.connections || []).map(convertConnection);
    
    const connectionCount = connections.length;
    
    if (writeMode) {
        // Ensure .vscode directory exists
        const vscodeDir = path.dirname(settingsPath);
        if (!fs.existsSync(vscodeDir)) {
            fs.mkdirSync(vscodeDir, { recursive: true });
        }
        
        // Read existing settings or start fresh
        let settings = {};
        if (fs.existsSync(settingsPath)) {
            settings = JSON.parse(fs.readFileSync(settingsPath, 'utf8'));
        }
        
        // Update sqltools.connections
        settings['sqltools.connections'] = connections;
        
        // Write back
        fs.writeFileSync(settingsPath, JSON.stringify(settings, null, 2));
        
        console.log(`✓ Updated ${settingsPath} with ${connectionCount} SQLTools connections`);
        console.log('\nConnections by group:');
        
        const byGroup = {};
        for (const conn of connections) {
            byGroup[conn.group] = (byGroup[conn.group] || 0) + 1;
        }
        for (const [group, count] of Object.entries(byGroup).sort()) {
            console.log(`  ${group}: ${count}`);
        }
        console.log('\nRequired extensions:');
        console.log('  - mtxr.sqltools (SQLTools)');
        console.log('  - mtxr.sqltools-driver-pg (PostgreSQL driver)');
        console.log('\n→ Install the extensions, then reload Cursor to see the connections');
    } else {
        // Print to stdout
        const settingsFormat = { "sqltools.connections": connections };
        console.log(JSON.stringify(settingsFormat, null, 2));
        
        console.error(`\nConverted ${connectionCount} connections.`);
        console.error(`\nTo write directly to .vscode/settings.json, run:`);
        console.error(`  node scripts/generate-sqltools-config.js --write`);
    }
}

main();



