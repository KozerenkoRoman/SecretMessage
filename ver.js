import path from 'path';
import fs from 'fs';
import { fileURLToPath } from 'url';
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const changeVersion = (file, version) => {
    if (!version) {
        console.error('Version is not defined');
        return;
    };
    const fullName = path.resolve(__dirname, file);
    if (!fs.existsSync(fullName)) {
        return;
    }
    try {
        let data;
        try {
            data = fs.readFileSync(fullName, { encoding: 'utf-8' });
        } catch (e) {
            console.error('File read failed', fullName, e.message);
            return null;
        };
        const regex = /("?version"?\s?[:=]\s?)(["']\d+.\d+.\d+["'])/ig;
        const updatedSample = data.replace(regex, `$1"${version}"`);
        console.info(`Updated file "${file}"`);
        fs.writeFileSync(file, updatedSample);
    } catch (err) {
        console.error(`Cannot read from file "${file}"`, err);
    }
};

const loadVersion = (file) => {
    const fullName = path.resolve(__dirname, file);
    if (!fs.existsSync(fullName)) {
        console.error(`File "${file}" does not exist`);
        return;
    };
    try {
        let data;
        try {
            data = fs.readFileSync(fullName, { encoding: 'utf-8' });
        } catch (e) {
            console.error('File read failed', fullName, e.message);
            return null;
        };
        const regex = /"?version"?\s?[:=]\s?["'](\d+.\d+.\d+)["']/ig;
        const version = regex.exec(data);
        if (!version) {
            return;
        }
        if (version.length > 0) {
            return version[1];
        }
    } catch (err) {
        console.error(`Cannot read from file "${file}"`, err);
    }
};

const incVersion = (version) => {
    if (!version) {
        return;
    }
    const arr = version.split('.');
    arr[arr.length - 1] = parseInt(arr[arr.length - 1]) + 1;
    return arr.join('.');
};

const version = incVersion(loadVersion('docker-build.ps1'));
console.info('Version:', version);
changeVersion('package.json', version);
changeVersion('swagger.json', version);
changeVersion('./src/common/constants.js', version);
changeVersion('docker-build.ps1', version);