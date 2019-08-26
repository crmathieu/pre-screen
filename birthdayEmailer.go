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

package main

import (
  "github.com/go-redis/redis"
  "time"
  "fmt"
  "strconv"
)

// implementation in GO
// My understanding of the initial script is that given a set of users whose id and birthdays are stored in redis,
// the emailer determines for each user if the current day correspond to the user's birthday and send a message accordingly.
//
// keys changes:
// - User struct now has a reference to a redis client variable instead of its own redis client variable
// - Redis keys have a prefix specific to this app to resolve possible conflict with data from other apps 
// - Removed save method from User struct (unused)
// 
// Notes:
// - local Redis is used for testing
// - birthdayEmailer_test.go contains the tests data set and code (run "go test")

const localhost = "127.0.0.1:6379"

type Cache struct {
  rc *redis.Client
}

const APP_PREFIX = "age-app:"

func createClient() (*Cache, error) {
   cache := Cache{redis.NewClient(&redis.Options{
                              Addr:     localhost,
                              Password: "",
                              DB:       0,
                            })}
    _, err := cache.rc.Ping().Result()
    return &cache, err          
}

type User struct {
  id int
  birthday time.Time
  c *Cache // the User struct now holds a pointer to a redis client, as opposed to a redis client 
}

func NewUser(id int, birthday time.Time, cache *Cache) *User {

  u := &User{}
  if id >= 0 {
    u.id = id
  } else {
    // this will not happen...
    u.id = int(cache.rc.Incr(APP_PREFIX+"user-ids").Val());
  }
  u.birthday = birthday
  u.c = cache
  return u
}


func (u *User)isBirthday() bool {
  d := time.Now()
  fmt.Printf("%v - %v\n", d, u.birthday)
  return u.birthday.Month() == d.Month() &&
         u.birthday.Day() == d.Day()
}  

func (u *User)CelebrateBirthday() {
  key := fmt.Sprintf(APP_PREFIX+"sent-%v", u.id)
  hasSent := u.hasSentThisYear(key)
  if !hasSent {
    u.SendBirthdayEmail()
  }
  u.setSentStatus(key)
}

func (u *User) hasSentThisYear(key string) bool {
  if u.c.rc.Get(key).Err() == nil {
    return true
  } 
  return false
}

func (u *User) SendBirthdayEmail() {
  fmt.Println("---> Sending Happy birthday to", u.id)
}

func (u *User) setSentStatus(key string) {
  oneYear := 60*60*24*365;
  u.c.rc.SetNX(key, u.id, time.Duration(oneYear)*time.Second) // set to expire in one year, we check later. 
}

func (u *User) save() {
  u.c.rc.HSet(APP_PREFIX+"user-"+strconv.Itoa(u.id), "birthday", u.birthday.String())
}


func checkBirthdays(c *Cache, id, maxid int) {
  if id > maxid {
    return
  }
  user := Find(c, id)
  if user != nil {
    if user.isBirthday() {
      user.CelebrateBirthday()
    }
  }
  checkBirthdays(c, id+1, maxid)
}

func Find(c *Cache, id int) *User {
  key := APP_PREFIX+"user-"+strconv.Itoa(id)
  result := c.rc.HGet(key, "birthday")

  if result.Err() == nil {
    t, err := time.Parse("2006-01-02", result.Val())
    if err == nil {
      return NewUser(id, t, c)
    } else {
      fmt.Println(id, result.Err())
    }
  } else {
    fmt.Println(id, result.Err())
  }
  return nil
}

func main() {
  if c, err := createClient(); err == nil {
    highestID, _ := strconv.Atoi(c.rc.Get(APP_PREFIX+"user-ids").Val())
    checkBirthdays(c, 0, highestID)
  } else {
    fmt.Println("Could not initialize Redis")
  }
}
