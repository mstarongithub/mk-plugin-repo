export function getAIscriptVersion(str: string): string | null {
    const regex = /^\/\/\/ @ (.*)/m;
    const match = str.match(regex);
    if (match && match.length > 1) {
        return match[1].trim();
    }
    return null;
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