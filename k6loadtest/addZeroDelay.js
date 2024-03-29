import http  from 'k6/http';
import exec from 'k6/execution';
import {check} from 'k6';

const baseUrl = 'http://127.0.0.1/'

export let options = {
    stages: [
        { duration:  '5s', target: 350 },
        { duration:  '4m', target: 350 },
    ],
};

export default function() {

    let delay = 5;
    let queue = 'qqqq'
    let body = exec.vu.idInTest + "." + exec.scenario.iterationInTest;

    let res = http.get(`${baseUrl}/add?queue=${queue}&body=${body}&delay=${delay}`, options);

    if (res.status !== 200) {
        console.log(res.status + " Err code: " + res.error_code)
    }

    check(res, {
        'status is 200': r => r.status === 200,
    });
};