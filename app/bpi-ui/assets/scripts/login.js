app.component('bpi-login', {
    template: `<template>
        <form @submit.prevent="submitForm">
            <div class="form-control">
                <label for="email">Email</label>
                <input type="email" id="email" v-model.trim="email"></input>
            </div>
            <div class="form-control">
                <label for="password">Password</label>
                <input type="password" id="password" v-model.trim="password"></input>
            </div>
            <button>Login</button>
        </form>
    </template>`,
    style: {},
    data() {
        return {
            email: '',
            password: '',
            formIsValid: true,
        };
    },
    methods: {
        submitForm() {
            axios.get('http://' + api_host + '/v1/users/token/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1', {
                auth: {
                    username: 'admin@example.com',
                    password: 'gophers'
                }
            })
            .then(res => {
                console.log(res);
            })
            .catch(err => {
                console.log(err);
            });
        }
    }
});
