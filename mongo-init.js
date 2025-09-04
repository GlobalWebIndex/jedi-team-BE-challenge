db.createUser(
    {
        user: "user-history-user",
        pwd: "password",
        roles: [
            {
                role: "readWrite",
                db: "user-history"
            }
        ]
    }
);


db.createCollection("userhistory", 
{
    validator: 
        {
            $jsonSchema: {
              required: [
                'sessionId'
              ],
              properties: {
                sessionId: {
                  bsonType: 'string',
                  description: '\'sessionId\' must be a string and is required'
                }

              }
        }
    }
 }
);