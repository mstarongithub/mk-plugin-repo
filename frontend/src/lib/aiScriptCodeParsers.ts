export function getAIscriptVersion(str: string): string | null {
    const regex = /^\/\/\/ @ (.*)/m;
    const match = str.match(regex);
    if (match && match.length > 1) {
        return match[1].trim();
    }
    return null;
}

export function getAIscriptPermissions(str: string): string[] | null {
    const regex = /###\s*{\s*.*permissions:\s*\[([^\]]*)\].*\s*}/s;
    const match = str.match(regex);
    if (match && match.length > 1) {
        return JSON.parse(`[${match[1].trim()}]`);
    }
    return null;
}

const PERMISSION_WARNINGS = {
    'write:pages': 'This plugin can modify your pages',
    'write:notifications': 'This plugin can send and modify your notifications',
    'write:reactions': 'This plugin can react to posts for you',
    'write:following': 'This plugin can follow people for you',
    'write:drive': 'This plugin can write to your drive',
    'read:account': 'This plugin can read account data',
    'write:account': 'This plugin can edit account data',
    'read:admin': 'This plugin can read admin data',
    'write:admin': 'This plugin can write admin data',

    // 'write:': "This plugin can write to your account",
}

export function getCodeWarnings(str : string) : string[] | null {
    const permissions = getAIscriptPermissions(str);

    if (!permissions) return [];

    const warnings = [];

    for (let i = 0; i < permissions.length; i++) {
        const permission = permissions[i];
        for (const warningKey in PERMISSION_WARNINGS) {
            if (permission.includes(warningKey)) {
                // eslint-disable-next-line @typescript-eslint/ban-ts-comment
                //@ts-expect-error
                warnings.push(PERMISSION_WARNINGS[warningKey]);
            }
        }
    }


    return warnings;
}

export function getPluginVersion(str: string): string | null {
    const regex = /###\s*{\s*.*version:\s*"([^"]*)".*\s*}/s;
    const match = str.match(regex);
    if (match && match.length > 1) {
        return match[1].trim();
    }
    return null;
}

export function getPluginName(str: string): string | null {
    const regex = /###\s*{\s*.*name:\s*"([^"]*)".*\s*}/s;
    const match = str.match(regex);
    if (match && match.length > 1) {
        return match[1].trim();
    }
    return null;
}

export function getPluginDesc(str: string): string | null {
    const regex = /###\s*{\s*.*description:\s*"([^"]*)".*\s*}/s;
    const match = str.match(regex);
    if (match && match.length > 1) {
        return match[1].trim();
    }
    return null;
}