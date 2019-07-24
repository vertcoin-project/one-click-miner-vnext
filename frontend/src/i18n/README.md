# Translation workflow

If you want to add a new translation to the Vertcoin OCM you can follow two tracks:

## Create a pull-request (preferable)

We expect you're familiar with how to create pull requests. If you're not, check [this article](https://akrabat.com/the-beginners-guide-to-contributing-to-a-github-project/)

### Step 1: Create a copy of the english base file

Once in your local fork branch, make a copy of `en.js` in this directory (`frontend/src/i18n`), and rename it to match your desired language (for instance, for German you'd rename it to `de.js`).

### Step 2: Translate!

Translate all the strings in the javascript file. Only translate the values not the identifiers, so:

```javascript
export default {
    generic : {
        "retry" : "Retry",
        "back_to_wallet" : "Back to wallet",
    },
    ...
}
```

Would become

```javascript
export default {
    generic : {
        "retry" : "Opnieuw proberen",
        "back_to_wallet" : "Terug naar portemonnee",
    },
    ...
}
```
**NOTE: Special characters**

There are a couple of special characters that are not allowed in javascript string literals, including backslashes and double quotes. You need to escape them. But since they're not used at all in the English base text, it seems unlikely you'll need them. In case of doubt, you can escape them [here](https://www.freeformatter.com/json-escape.html)

### Step 3: Add language to frontend

In the file `frontend/src/main.js` there's a list of the translations imported - add your new language there:

```javascript
// Import all locales
import locale_en from "./i18n/en.js";
import locale_nl from "./i18n/nl.js";
import locale_de from "./i18n/de.js"; // <-- this line is added
```

Further down in the file, also add it to the list of languages injected to the i18n component:

```javascript
    const i18n = new VueI18n({
      locale: result, // set locale
      messages : {
        en: locale_en,
        nl: locale_nl,
        de: locale_de, // <-- this line is added
      },
    });
```

### Step 4: Add language to the backend

The host code running on the machine does the detection of the language and chooses the most appropriate one based on the user's locale. It needs to be made aware of the newly available language. Add this to the file `backend/languages.go` around line 9:

```golang
var availableLanguages = []string{
	"en",
    "nl",
    "de", // <-- this line is added. Notice the comma on the end - it belongs there!
}
```

### Step 5: Commit & Create PR

You're done. Add all files to a commit, push it to your personal fork and then create a pull request to the main OCM repository. Thanks a ton for your contribution in advance!

## Alternative option (only if you fail at the above)

The alternative option is that you just download the `en.js` file, translate it locally, and open an issue including the translated file. Then I can include it for you. But if you're able to use the PR workflow, I would really appreciate and prefer you use that!