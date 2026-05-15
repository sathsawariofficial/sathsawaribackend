import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';
import { check, sleep } from 'k6';
import http from 'k6/http';

// Test Configuration
export const options = {
    scenarios: {
        user_journey: {
            executor: 'constant-arrival-rate',
            rate: 200, // 100 requests per second
            timeUnit: '1s', // per second
            duration: '2m', // total test duration
            preAllocatedVUs: 50, // initial pool of virtual users
            maxVUs: 100, // maximum virtual users
        },
    },
};

// Utility function to generate random data
function generateDriverData() {
    const mobile = `+92${Math.floor(3000000000 + Math.random() * 9000000000)}`; // Random Pakistani mobile number
    const name = `Driver_${uuidv4().substring(0, 6)}`;
    const password = `Pass_${uuidv4().substring(0, 6)}`;
    const vehicleNumber = `CAR${Math.floor(1000 + Math.random() * 9000)}`;
    const vehicleInfo = `Model_${Math.floor(2000 + Math.random() * 23)}`;
    return { mobile, name, password, vehicleNumber, vehicleInfo };
}

// Test Flow for Each User
export default function () {
    const driver = generateDriverData();

    // // Step 1: Register Driver
    // let res = http.post(
    //     'https://sawarilink.com/api/v1/driver/register',
    //     JSON.stringify({
    //         mobile: driver.mobile,
    //         name: driver.name,
    //         password: driver.password,
    //         vehicle_number: driver.vehicleNumber,
    //         vehicle_info: driver.vehicleInfo,
    //     }),
    //     {
    //         headers: { 'Content-Type': 'application/json' },
    //     }
    // );

    // check(res, {
    //     'Register status is 200': (r) => r.status === 200,
    // });

    // // Extract registration response if needed
    // sleep(1);

    // Step 2: Login Driver
    let res = http.post(
        'https://sawarilink.com/api/v1/driver/login',
        JSON.stringify({
            mobile: "03405421037",
            password: "Golang@12122",
        }),
        {
            headers: { 'Content-Type': 'application/json' },
        }
    );

    const session_id = res.json('session_id');

    check(res, {
        'Login status is 200': (r) => r.status === 200,
        'Session ID captured': () => session_id !== undefined,
    });

    sleep(1);

    // Step 3: Create Ride
    res = http.post(
        'https://sawarilink.com/api/v1/ride/create',
        JSON.stringify({
            start_location: 'Point A',
            end_location: 'Point B',
            start_datetime: '2024-12-06 10:00:00',
            estimated_end_datetime: '2024-12-07 11:00:00',
            number_of_seats: 3,
            fare: 15.0,
            route_details: 'Via Main Street',
        }),
        {
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${session_id}`,
            },
        }
    );

    check(res, {
        'Create ride status is 200': (r) => r.status === 200,
    });

    sleep(1);

    // Step 4: Get Rides
    res = http.get(
        'https://sawarilink.com/api/v1/ride/filtered',
        {
            params: {
                page: 1,
                start_location: 'Point A',
                end_location: 'Point B',
                start_datetime: '2024-12-06 10:00:00',
                estimated_end_datetime: '2024-12-07 11:00:00',
            },
            headers: {
                'Content-Type': 'application/json',
            },
        }
    );

    check(res, {
        'Get rides status is 200': (r) => r.status === 200,
    });

    sleep(1); // Simulate think time
}
