"use strict";
var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
var __generator = (this && this.__generator) || function (thisArg, body) {
    var _ = { label: 0, sent: function() { if (t[0] & 1) throw t[1]; return t[1]; }, trys: [], ops: [] }, f, y, t, g;
    return g = { next: verb(0), "throw": verb(1), "return": verb(2) }, typeof Symbol === "function" && (g[Symbol.iterator] = function() { return this; }), g;
    function verb(n) { return function (v) { return step([n, v]); }; }
    function step(op) {
        if (f) throw new TypeError("Generator is already executing.");
        while (_) try {
            if (f = 1, y && (t = op[0] & 2 ? y["return"] : op[0] ? y["throw"] || ((t = y["return"]) && t.call(y), 0) : y.next) && !(t = t.call(y, op[1])).done) return t;
            if (y = 0, t) op = [op[0] & 2, t.value];
            switch (op[0]) {
                case 0: case 1: t = op; break;
                case 4: _.label++; return { value: op[1], done: false };
                case 5: _.label++; y = op[1]; op = [0]; continue;
                case 7: op = _.ops.pop(); _.trys.pop(); continue;
                default:
                    if (!(t = _.trys, t = t.length > 0 && t[t.length - 1]) && (op[0] === 6 || op[0] === 2)) { _ = 0; continue; }
                    if (op[0] === 3 && (!t || (op[1] > t[0] && op[1] < t[3]))) { _.label = op[1]; break; }
                    if (op[0] === 6 && _.label < t[1]) { _.label = t[1]; t = op; break; }
                    if (t && _.label < t[2]) { _.label = t[2]; _.ops.push(op); break; }
                    if (t[2]) _.ops.pop();
                    _.trys.pop(); continue;
            }
            op = body.call(thisArg, _);
        } catch (e) { op = [6, e]; y = 0; } finally { f = t = 0; }
        if (op[0] & 5) throw op[1]; return { value: op[0] ? op[1] : void 0, done: true };
    }
};
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
// import commands from "probot-commands";
var scathach_api_1 = __importDefault(require("scathach-api"));
var config = {
    useragent: '',
    apikey: ''
};
var sca = new scathach_api_1.default(config);
module.exports = function (app) {
    app.log("Im alive!");
    app.on("issues.opened", function (context) { return __awaiter(void 0, void 0, void 0, function () {
        var list, _a, img, thanks, issueComment;
        return __generator(this, function (_b) {
            switch (_b.label) {
                case 0: return [4 /*yield*/, sca.getReaction('uwu')];
                case 1:
                    _a = [_b.sent()];
                    return [4 /*yield*/, sca.getReaction('uwu')];
                case 2:
                    _a = _a.concat([_b.sent()]);
                    return [4 /*yield*/, sca.getReaction('pout')];
                case 3:
                    _a = _a.concat([_b.sent()]);
                    return [4 /*yield*/, sca.getReaction('smug')];
                case 4:
                    list = _a.concat([_b.sent()]);
                    img = list[Math.floor(Math.random() * list.length)];
                    thanks = [
                        "You'll get a response soon! UwU",
                        "Thanks for opening this issue, we will get back to you soon! OwO",
                        "I see you opened an issue, an admins or collaborators will get back to you soon! UmU",
                    ];
                    issueComment = context.issue({
                        //mention username
                        body: "Hey! @".concat(context.payload.issue.user.login, " ").concat(thanks[Math.floor(Math.random() * thanks.length)], "\n      <details>\n      <summary>Click here to make your day UmU</summary>\n      ![abc](").concat(img.url, " \"UmU\")\n      </details>"),
                    });
                    return [4 /*yield*/, context.octokit.issues.createComment(issueComment)];
                case 5:
                    _b.sent();
                    return [4 /*yield*/, context.octokit.issues.addLabels(context.issue({ labels: ["triage"] }))];
                case 6:
                    _b.sent();
                    return [2 /*return*/];
            }
        });
    }); });
    // create event pull request
    app.on("pull_request.opened", function (context) { return __awaiter(void 0, void 0, void 0, function () {
        var img, thanks, issueComment;
        return __generator(this, function (_a) {
            switch (_a.label) {
                case 0: return [4 /*yield*/, sca.getReaction('pat')];
                case 1:
                    img = _a.sent();
                    thanks = [
                        "Thanks for making a pull request, an collaborators will get back to you soon! UwU",
                        "We will get back to you soon! OwO",
                        "What a great improvements, we will consider it! UmU",
                    ];
                    issueComment = context.issue({
                        //mention username
                        body: "Hey?! @".concat(context.payload.pull_request.user.login, " ").concat(thanks[Math.floor(Math.random() * thanks.length)], "\n      <details>\n      <summary>Click here to make your day UmU</summary>\n      ![abc](").concat(img.url, " \"UmU\")\n      </details>")
                    });
                    return [4 /*yield*/, context.octokit.issues.createComment(issueComment)];
                case 2:
                    _a.sent();
                    return [4 /*yield*/, context.octokit.issues.addLabels(context.issue({ labels: ["pr:pending"] }))];
                case 3:
                    _a.sent();
                    return [2 /*return*/];
            }
        });
    }); });
    /*
    commands(app, "label", async (context: any, command: any) => {
      console.log(context.payload.comment.user);
      const labels = ##
      await context.octokit.issues.addLabels(context.issue({ labels }));
  
      
      const issueComment = context.issue({
        //mention username
        body: `welp`
      });
      await context.octokit.issues.createComment(issueComment);
     
    });
   */
};
//# sourceMappingURL=index.js.map