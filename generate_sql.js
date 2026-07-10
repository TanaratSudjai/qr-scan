const fs = require('fs');
const file = '/Users/tanarat/Desktop/qr-scan/data.seed.json';
const data = JSON.parse(fs.readFileSync(file, 'utf8'));

let sql = `CREATE TABLE IF NOT EXISTS member_ai (
    id INT AUTO_INCREMENT PRIMARY KEY,
    timestamp DATETIME,
    fullName VARCHAR(255),
    address TEXT,
    province VARCHAR(100),
    postalCode VARCHAR(20),
    phoneNumber VARCHAR(50),
    email VARCHAR(255),
    organization VARCHAR(255),
    position VARCHAR(255),
    responsibility TEXT,
    expectation TEXT
);\n\n`;

sql += `INSERT INTO member_ai (timestamp, fullName, address, province, postalCode, phoneNumber, email, organization, position, responsibility, expectation) VALUES\n`;

const values = data.map(item => {
    const escape = (str) => {
        if (str === null || str === undefined) return 'NULL';
        return "'" + String(str).replace(/'/g, "''").replace(/\n/g, '\\n') + "'";
    };
    return `(${escape(item.timestamp)}, ${escape(item.fullName || item['ชื่อ-sgn'])}, ${escape(item.address)}, ${escape(item.province)}, ${escape(item.postalCode)}, ${escape(item.phoneNumber)}, ${escape(item.email)}, ${escape(item.organization)}, ${escape(item.position)}, ${escape(item.responsibility)}, ${escape(item.expectation)})`;
});

sql += values.join(',\n') + ';';

fs.writeFileSync('/Users/tanarat/Desktop/qr-scan/seed_member_ai.sql', sql, 'utf8');
console.log('Done');
