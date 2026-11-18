. as $s | ("microjaas_dex_" + (.uuid | gsub("-"; ""))) as $db | {
  "issuer": "https://microjaas.test:18556/dex",
  "storage": {
    "type": "postgres",
    "config": {
      "host": $host,
      "port": $port,
      "database": $db,
      "user": $db,
      "password": $s.database_passwords.dex,
      "ssl": {"mode": "disable"}
    }
  },
  "web": {
    "https": "microjaas.test:18556",
    "tlsCert": ($common + "/config/tls.crt"),
    "tlsKey": ($common + "/config/tls.key")
  },
  "frontend": {
    "issuer": "MicroJAAS",
    "dir": ($root + "/share/dex/web")
  },
  "oauth2": {
    "skipApprovalScreen": true,
    "grantTypes": ["authorization_code", "refresh_token", "urn:ietf:params:oauth:grant-type:device_code", "client_credentials"]
  },
  "staticClients": [{
    "id": "microjaas",
    "name": "MicroJAAS",
    "secret": $s.oauth_secret,
    "redirectURIs": ["https://microjaas.test:18443/auth/callback", "/dex/device/callback"]
  }, {
    "id": "automation",
    "name": "Automation",
    "secret": "microjaas-development-only"
  }],
  "enablePasswordDB": true,
  "staticPasswords": [{
    "email": "admin@microjaas.local",
    "hash": $s.admin_hash,
    "username": "admin",
    "userID": $s.uuid
  }]
}
