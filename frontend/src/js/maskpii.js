// Mask sensitive information (IPs, MACs, emails, domains, secrets, user accounts) in logs
export function maskPII(logStr) {
  if (!logStr) return "";
  let res = logStr;

  const ipMap = new Map();
  const macMap = new Map();
  const emailMap = new Map();
  const hostMap = new Map();
  const userMap = new Map();

  // 1. Mask Secrets/Passwords/Tokens
  res = res.replace(/(password|passwd|pwd|token|api_?key|secret|auth|bearer|private_?key)\s*([:=])\s*([^\s,;]+)/gi, '$1$2[REDACTED]');

  // 2. Mask Emails
  res = res.replace(/\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b/g, (m) => {
    if (!emailMap.has(m)) {
      emailMap.set(m, `[EMAIL_${emailMap.size + 1}]`);
    }
    return emailMap.get(m);
  });

  // 3. Mask MAC addresses
  res = res.replace(/\b(?:[0-9A-Fa-f]{2}[:-]){5}[0-9A-Fa-f]{2}\b/g, (m) => {
    if (!macMap.has(m)) {
      macMap.set(m, `[MAC_${macMap.size + 1}]`);
    }
    return macMap.get(m);
  });

  // 4. Mask IPv4 addresses
  res = res.replace(/\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b/g, (m) => {
    if (!ipMap.has(m)) {
      ipMap.set(m, `[IP_${ipMap.size + 1}]`);
    }
    return ipMap.get(m);
  });

  // 5. Mask IPv6 addresses
  res = res.replace(/\b(?:[0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}\b/g, (m) => {
    if (!ipMap.has(m)) {
      ipMap.set(m, `[IP_${ipMap.size + 1}]`);
    }
    return ipMap.get(m);
  });

  // 6. Mask Domains/Hostnames
  res = res.replace(/\b(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+(?:com|net|org|edu|gov|io|co|jp|cn|de|uk|info|biz|me|xyz|tech|online|dev|site|internal|local|corp|domain)\b/g, (m) => {
    if (!hostMap.has(m)) {
      hostMap.set(m, `[HOST_${hostMap.size + 1}]`);
    }
    return hostMap.get(m);
  });

  // 7. Mask Users in common patterns: e.g. "user <name>", "user=<name>", "for user <name>", "from user <name>"
  res = res.replace(/\b(user(?:name)?|for user|from user)\s*([:=]|\s+)\s*([a-zA-Z0-9._-]+)/gi, (m, prefix, sep, uname) => {
    if (uname === '[REDACTED]' || uname.startsWith('[')) {
      return m;
    }
    if (!userMap.has(uname)) {
      userMap.set(uname, `[USER_${userMap.size + 1}]`);
    }
    return `${prefix}${sep}${userMap.get(uname)}`;
  });

  return res;
}
