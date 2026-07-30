import { createServer } from "node:http";

const host = "127.0.0.1";
const port = 3101;

const user = {
  id: "user-1",
  role_id: "role-1",
  name: "Alex Morgan",
  bio: "Fleet operator",
  username: "alex",
  preferences: {},
  created_at: "2026-07-30T00:00:00Z",
};

const permissionNames = [
  "profile:get",
  "profile:set",
  "profile_security:set",
  "node:get",
  "node_log:get",
];

function sendJson(response, status, body) {
  response.writeHead(status, { "Content-Type": "application/json" });
  response.end(JSON.stringify(body));
}

const server = createServer((request, response) => {
  if (request.url === "/health") {
    response.writeHead(204);
    response.end();
    return;
  }

  if (!request.headers.authorization?.startsWith("Bearer ")) {
    sendJson(response, 401, {
      error: "Unauthorized",
      message: "Bearer token required",
    });
    return;
  }

  if (request.method === "GET" && request.url === "/api/v1/profile") {
    sendJson(response, 200, user);
    return;
  }

  if (
    request.method === "GET" &&
    request.url === "/api/v1/profile/permissions"
  ) {
    sendJson(
      response,
      200,
      permissionNames.map((name, index) => ({
        id: `permission-${index + 1}`,
        name,
        description: `${name} fixture`,
        preferences: {},
        created_at: "2026-07-30T00:00:00Z",
      })),
    );
    return;
  }

  sendJson(response, 404, {
    error: "Not Found",
    message: `No deterministic fixture for ${request.method} ${request.url}`,
  });
});

server.listen(port, host);
