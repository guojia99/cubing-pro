import * as crypto from 'crypto';

// 加密函数
function encrypt(plaintext: string, key: Buffer): string {
    const iv = crypto.randomBytes(16);
    const cipher = crypto.createCipheriv('aes-256-cbc', key, iv);

    let encrypted = cipher.update(plaintext, 'utf8', 'base64');
    encrypted += cipher.final('base64');
    return iv.toString('base64') + ':' + encrypted;
}

// 解密函数
function decrypt(ciphertext: string, key: Buffer): string {
    const [ivStr, encrypted] = ciphertext.split(':');
    const iv = Buffer.from(ivStr, 'base64');
    const decipher = crypto.createDecipheriv('aes-256-cbc', key, iv);

    let decrypted = decipher.update(encrypted, 'base64', 'utf8');
    decrypted += decipher.final('utf8');
    return decrypted;
}

// 主函数
function main() {
    const key = Buffer.from('your-key', 'utf8'); // 设置你的密钥
    const plaintext = 'Hello, World!';

    // 加密
    const encrypted = encrypt(plaintext, key);
    console.log('Encrypted:', encrypted);

    // 解密
    const decrypted = decrypt(encrypted, key);
    console.log('Decrypted:', decrypted);
}

main();


// 生成长度为32的随机字符串
function generateRandomKey(timestamp: number): string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    let key = '';
    const random = (seed: number) => {
        const x = Math.sin(seed) * 10000;
        return x - Math.floor(x);
    };

    for (let i = 0; i < 32; i++) {
        const seed = timestamp + i;
        const idx = Math.floor(random(seed) * charset.length);
        key += charset.charAt(idx);
    }

    return key;
}

// 获取当前时间戳
const timestamp = Math.floor(new Date().getTime() / 1000);

// 生成随机字符串
const randomKey = generateRandomKey(timestamp);
