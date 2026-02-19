import * as crypto from 'crypto';

export function computeHash(content: string): string {
    return crypto.createHash('md5').update(content).digest('hex');
}

export class HashCache<T> {
    private cache: Map<string, { hash: string; data: T }> = new Map();

    set(key: string, hash: string, data: T): void {
        this.cache.set(key, { hash, data });
    }

    get(key: string, hash: string): T | undefined {
        const cached = this.cache.get(key);
        if (cached && cached.hash === hash) {
            return cached.data;
        }
        return undefined;
    }

    has(key: string, hash: string): boolean {
        const cached = this.cache.get(key);
        return cached !== undefined && cached.hash === hash;
    }

    delete(key: string): boolean {
        return this.cache.delete(key);
    }

    clear(): void {
        this.cache.clear();
    }

    size(): number {
        return this.cache.size;
    }
}
