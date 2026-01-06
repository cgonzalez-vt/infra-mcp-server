#!/usr/bin/env node
/**
 * Exports SQLTools configuration for sharing with other developers.
 * Passwords are replaced with a placeholder - devs need to get the password separately.
 * 
 * Usage:
 *   node scripts/export-sqltools-config.js > sqltools-connections.json
 *   node scripts/export-sqltools-config.js --with-passwords  # Include actual passwords (be careful!)
 */

const fs = require('fs');
const path = require('path');

const args = process.argv.slice(2);
const includePasswords = args.includes('--with-passwords');
const configPath = path.join(__dirname, '../bin/config.json');

function convertConnection(conn, includePasswords) {
    const driver = conn.type === 'postgres' ? 'PostgreSQL' : 'MySQL';
    
    const connection = {
        name: conn.display_name || conn.id,
        driver: driver,
        server: conn.host,
        port: conn.port,
        database: conn.name,
        username: conn.user,
        password: includePasswords ? conn.password : "ASK_TEAM_LEAD_FOR_PASSWORD",
        group: conn.environment || 'default'
    };
    
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
    
    const connections = (config.connections || []).map(conn => convertConnection(conn, includePasswords));
    
    const output = {
        "sqltools.connections": connections
    };
    
    console.log(JSON.stringify(output, null, 2));
    
    if (!includePasswords) {
        console.error('\n---');
        console.error('NOTE: Passwords replaced with placeholder.');
        console.error('To include passwords: node scripts/export-sqltools-config.js --with-passwords');
        console.error('\nInstructions for devs:');
        console.error('1. Install extensions: mtxr.sqltools and mtxr.sqltools-driver-pg');
        console.error('2. Copy the JSON above into .vscode/settings.json (or user settings)');
        console.error('3. Replace "ASK_TEAM_LEAD_FOR_PASSWORD" with the actual password');
        console.error('4. Reload VS Code / Cursor');
    }
}

main();

