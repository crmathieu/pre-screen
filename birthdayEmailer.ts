// NOTE TO THE READER:
// The code below is written to be "typescript-esque" but you can consider this to be pseudo-code for this evaluation.
// Your rewrite of this code can be in any language you like. Dont get hung-up on language specific problems or linting issues.
// * We suggest either Python or Javascript, as that is what we use here at Dragonchain.

// EXPLINATION:
// A Junior developer is asking for your help!
// Read this file completely then identify what the junior developer is trying to do and refactor it to make more sense.
// Feel free to add new files/folders if you wish, but lets try to keep it as small as possible.
// Leave comments explaining your reasoning, and why you chose to refactor things the way you did.

// INSTRUCTIONS:
// 1. clone this repo.
// 1. refactor this locally as you see fit.
// 1. push your refactored code to your own github account.
// 1. send a link to your new repo via EMAIL for someone to review.

// CONSIDERATIONS:
// 1. this app uses a shared redis with other apps.
// 1. a cron job will run once a day and execute the main function.

// To help guide your refactor:
// - what are the problems at scale?
// - what edge cases did this dev not think of?
// - any issues with consistency, naming conventions?
// - what here do you disagree with?
// - what best-practices are being violated here?


// Extra Credit:
// - refactor this OOP approach to use the functional programming style.
// - prove it works with TESTS!

import {createClient} from 'redis'; // <-- Let's' pretend this is the standard NodeJS redis package.
import {sendBirthdayEmail} from './lib'; // <-- Let's pretend this is already a thing that works.

class User {
  id: number;
  birthday: Date;
  redis: any;

  public constructor(id?: number, birthday?: Date){
    this.id = id || this.redis.incr('age-app:user-ids');
    this.birthday = birthday;
    this.redis = createClient();
  }

  public static async find(id){
    const user = await new User().redis.hget('users', id);
    return new User(user.id, new Date(user.birthday));
  }

  public isBirthday() {
    const d = new Date();
    return this.birthday.getFullYear() === d.getFullYear() &&
      this.birthday.getMonth() === d.getMonth() &&
      this.birthday.getDate() === d.getDate();
  }

  public async celebrateBirthday() {
    const key = `sent-${this.id}`;
    const hasSent = await this.hasSentThisYear(key);
    if(!hasSent) await sendBirthdayEmail();
    await this.setSentStatus(key);
    this.save();
  }

  public hasSentThisYear(key: string) {
    return this.redis.get(key) ? true : false;
  }
  
  public setSentStatus(key: string) {
    const oneYear = 60*60*24*365;
    this.redis.setex(key, oneYear); // set to expire in one year, we check later. 
  }

  public save(){
    this.redis.hset('users', this.id, 'birthday', this.birthday.toISOString());
  }
}

// Main function (run once a day, every day)
export const main = async () => {
  const r = createClient()
  const highestId = await r.get('age-app:user-ids')
  for(var i = 0; i <= highestId; i++) {
    const user = await User.find(i);
    if(user.isBirthday()) await user.celebrateBirthday();
  }
};
