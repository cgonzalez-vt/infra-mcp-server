#!/usr/bin/env node
/**
 * Converts infra-mcp-server config.json to Database Client (cweijan) VS Code extension format
 * and writes directly to the extension's SQLite storage.
 * 
 * Usage:
 *   node scripts/generate-dbclient-config.js [options] [config-file]
 * 
 * Options:
 *   --write    Write directly to Cursor's Database Client storage
 *   --print    Print JSON to stdout (default)
 * 
 * Examples:
 *   node scripts/generate-dbclient-config.js                     # Print to stdout
 *   node scripts/generate-dbclient-config.js --write             # Write to extension storage
 *   node scripts/generate-dbclient-config.js bin/config.json     # Use custom config file
 */

const fs = require('fs');
const path = require('path');
const { execSync, spawnSync } = require('child_process');
const os = require('os');

const args = process.argv.slice(2);
const writeMode = args.includes('--write');
const configArg = args.find(arg => !arg.startsWith('--'));
const configPath = configArg || path.join(__dirname, '../bin/config.json');

// Path to Cursor's SQLite database
const dbPath = path.join(
    process.env.HOME,
    '.config/Cursor/User/globalStorage/state.vscdb'
);

function generateConnectionKey() {
    return Date.now().toString();
}

function convertConnection(conn, key) {
    const dbType = conn.type === 'postgres' ? 'PostgreSQL' : 'MySQL';
    
    return {
        host: conn.host,
        port: conn.port,
        user: conn.user,
        password: conn.password,  // Extension will encrypt on first use
        dbType: dbType,
        database: conn.name,
        name: conn.display_name || conn.id,
        group: conn.environment || 'default',
        advance: {
            idleConfig: { enable: true },
            hideSystemSchema: true,
            groupingTables: false,
            loadMetaDataWhenExpandTreeView: true
        },
        treeFeatures: [],
        usingSSH: false,
        useSocksProxy: false,
        useHTTPProxy: false,
        ssh: {
            host: "",
            port: 22,
            username: "root",
            type: "auto",
            privateKeyPath: "",
            serverType: "linux",
            connectTimeout: 5000,
            algorithms: {
                kex: [
                    "ecdh-sha2-nistp256",
                    "ecdh-sha2-nistp384",
                    "ecdh-sha2-nistp521",
                    "diffie-hellman-group-exchange-sha256",
                    "diffie-hellman-group14-sha256",
                    "diffie-hellman-group15-sha512",
                    "diffie-hellman-group16-sha512",
                    "diffie-hellman-group17-sha512",
                    "diffie-hellman-group18-sha512",
                    "diffie-hellman-group14-sha1",
                    "diffie-hellman-group-exchange-sha1"
                ],
                cipher: [
                    "aes128-gcm@openssh.com",
                    "aes256-gcm@openssh.com",
                    "aes128-ctr",
                    "aes192-ctr",
                    "aes256-ctr"
                ],
                serverHostKey: [
                    "ssh-ed25519",
                    "ecdsa-sha2-nistp256",
                    "ecdsa-sha2-nistp384",
                    "ecdsa-sha2-nistp521",
                    "rsa-sha2-512",
                    "rsa-sha2-256",
                    "ssh-rsa",
                    "ssh-dss"
                ]
            },
            disableSFTP: false,
            pruneSFTPRoot: true
        },
        global: true,
        savePassword: "Forever",
        readonly: false,
        sort: 11,
        useSSL: false,
        fs: {
            encoding: "utf8",
            showHidden: true
        },
        key: key,
        connectionKey: "database.connections"
    };
}

function getCurrentExtensionData() {
    try {
        const result = execSync(
            `sqlite3 "${dbPath}" "SELECT value FROM ItemTable WHERE key = 'cweijan.vscode-mysql-client2';"`,
            { encoding: 'utf8' }
        ).trim();
        return result ? JSON.parse(result) : null;
    } catch (e) {
        return null;
    }
}

function writeExtensionData(data) {
    // Use a SQL file to avoid escaping issues
    const sqlFile = path.join(os.tmpdir(), 'dbclient-update.sql');
    const jsonContent = JSON.stringify(data);
    
    // Escape single quotes for SQLite by doubling them
    const escapedJson = jsonContent.replace(/'/g, "''");
    
    fs.writeFileSync(sqlFile, `UPDATE ItemTable SET value = '${escapedJson}' WHERE key = 'cweijan.vscode-mysql-client2';`);
    
    // Execute the SQL file using stdin
    const result = spawnSync('sqlite3', [dbPath], {
        input: fs.readFileSync(sqlFile, 'utf8'),
        encoding: 'utf8'
    });
    
    // Clean up temp file
    fs.unlinkSync(sqlFile);
    
    if (result.status !== 0) {
        throw new Error(`SQLite error: ${result.stderr}`);
    }
}

function main() {
    if (!fs.existsSync(configPath)) {
        console.error(`Config file not found: ${configPath}`);
        process.exit(1);
    }
    
    const configData = fs.readFileSync(configPath, 'utf8');
    const config = JSON.parse(configData);
    
    // Convert all connections
    const connections = {};
    let sortOrder = 1;
    const baseKey = Date.now();
    
    for (const conn of config.connections || []) {
        const key = (baseKey + sortOrder).toString();
        const converted = convertConnection(conn, key);
        converted.sort = sortOrder++;
        connections[key] = converted;
    }
    
    const connectionCount = Object.keys(connections).length;
    
    if (writeMode) {
        if (!fs.existsSync(dbPath)) {
            console.error(`Cursor database not found at: ${dbPath}`);
            console.error('Make sure Cursor is installed and has been opened at least once.');
            process.exit(1);
        }
        
        // Get current extension data
        let extData = getCurrentExtensionData();
        
        if (!extData) {
            extData = {
                databaseMetas: null,
                "database.connections": {}
            };
        }
        
        // Merge new connections (keeping existing ones)
        extData["database.connections"] = {
            ...extData["database.connections"],
            ...connections
        };
        
        // Write back
        writeExtensionData(extData);
        
        console.log(`✓ Added ${connectionCount} database connections to Database Client extension`);
        console.log('\nConnections by group:');
        
        const byGroup = {};
        for (const conn of Object.values(connections)) {
            byGroup[conn.group] = (byGroup[conn.group] || 0) + 1;
        }
        for (const [group, count] of Object.entries(byGroup).sort()) {
            console.log(`  ${group}: ${count}`);
        }
        console.log('\n→ Reload Cursor window (Ctrl+Shift+P → "Developer: Reload Window") to see the connections');
    } else {
        // Print to stdout
        console.log(JSON.stringify({ "database.connections": connections }, null, 2));
        
        console.error(`\nConverted ${connectionCount} connections.`);
        console.error(`\nTo write directly to Database Client extension, run:`);
        console.error(`  node scripts/generate-dbclient-config.js --write`);
    }
}

main();
