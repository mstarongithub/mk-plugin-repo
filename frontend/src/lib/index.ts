// place files you want to import through the `$lib` alias in this folder.

export type Plugin = {
    id: number;
    name: string;
    summary_short: string;
    summary_long: string;
    current_version: string;
    all_versions: string[];
    tags: string[];
    author_id: number;
    type: 'plugin';
};