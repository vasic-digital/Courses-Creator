# HTML and CSS Fundamentals

Learn the building blocks of modern web development. This course covers HTML structure and CSS styling to create beautiful, responsive websites.

## Introduction to HTML

HTML (HyperText Markup Language) is the standard markup language for creating web pages. It provides the structure and content of web documents.

### What is HTML?
- **Markup Language**: Uses tags to define content structure
- **Standard**: Maintained by W3C (World Wide Web Consortium)
- **Version**: Currently HTML5 (released 2014)
- **Purpose**: Structure content, not presentation

### Basic HTML Document Structure
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>My First Web Page</title>
</head>
<body>
    <h1>Welcome to My Website</h1>
    <p>This is a paragraph of text.</p>
</body>
</html>
```

## HTML Elements and Tags

HTML uses elements (tags) to define different types of content.

### Headings
```html
<h1>Main Heading (most important)</h1>
<h2>Subheading</h2>
<h3>Sub-subheading</h3>
<h4>Level 4 heading</h4>
<h5>Level 5 heading</h5>
<h6>Least important heading</h6>
```

### Paragraphs and Text
```html
<p>This is a paragraph of text.</p>

<!-- Line breaks -->
<p>First line<br>Second line</p>

<!-- Horizontal rule -->
<hr>

<!-- Preformatted text (preserves whitespace) -->
<pre>
    This text
    maintains its
    formatting
</pre>
```

### Text Formatting
```html
<!-- Bold text -->
<strong>Important text</strong>
<b>Also bold (less semantic)</b>

<!-- Italic text -->
<em>Emphasized text</em>
<i>Also italic (less semantic)</i>

<!-- Underline -->
<u>Underlined text</u>

<!-- Strikethrough -->
<s>Strikethrough text</s>
<del>Deleted text</del>

<!-- Subscript and superscript -->
H<sub>2</sub>O
E = mc<sup>2</sup>

<!-- Code -->
<code>console.log('Hello World');</code>

<!-- Keyboard input -->
Press <kbd>Ctrl</kbd> + <kbd>C</kbd> to copy

<!-- Sample output -->
<samp>File not found</samp>

<!-- Variable -->
<var>x</var> = 42

<!-- Abbreviation -->
<abbr title="HyperText Markup Language">HTML</abbr>

<!-- Citation -->
<cite>The Great Gatsby</cite>

<!-- Quote -->
<blockquote>
    "To be or not to be, that is the question."
    <cite>William Shakespeare</cite>
</blockquote>

<!-- Short quote -->
<p>As Shakespeare said, <q>to thine own self be true</q>.</p>
```

### Lists

#### Unordered Lists
```html
<ul>
    <li>Apples</li>
    <li>Bananas</li>
    <li>Oranges</li>
</ul>
```

#### Ordered Lists
```html
<ol>
    <li>First item</li>
    <li>Second item</li>
    <li>Third item</li>
</ol>

<!-- With custom numbering -->
<ol start="10">
    <li>Item 10</li>
    <li>Item 11</li>
</ol>

<!-- With different numbering types -->
<ol type="A">
    <li>Item A</li>
    <li>Item B</li>
</ol>
```

#### Definition Lists
```html
<dl>
    <dt>HTML</dt>
    <dd>HyperText Markup Language</dd>
    <dt>CSS</dt>
    <dd>Cascading Style Sheets</dd>
</dl>
```

### Links and Navigation
```html
<!-- Basic link -->
<a href="https://www.example.com">Visit Example</a>

<!-- Link to email -->
<a href="mailto:contact@example.com">Email us</a>

<!-- Link to phone -->
<a href="tel:+1234567890">Call us</a>

<!-- Link to file -->
<a href="documents/resume.pdf" download>Download Resume</a>

<!-- Link to section on same page -->
<a href="#section1">Jump to Section 1</a>

<!-- Target attributes -->
<a href="https://example.com" target="_blank">Opens in new tab</a>
<a href="https://example.com" target="_self">Opens in same window</a>

<!-- Link with title -->
<a href="https://example.com" title="Visit our website">Our Site</a>
```

### Images
```html
<!-- Basic image -->
<img src="images/photo.jpg" alt="A beautiful landscape">

<!-- Image with dimensions -->
<img src="logo.png" alt="Company Logo" width="200" height="100">

<!-- Image with title -->
<img src="chart.png" alt="Sales Chart" title="Monthly Sales Data">

<!-- Responsive image -->
<img src="responsive.jpg" alt="Responsive Image" style="max-width: 100%; height: auto;">

<!-- Image as link -->
<a href="https://example.com">
    <img src="banner.jpg" alt="Click here">
</a>

<!-- Figure with caption -->
<figure>
    <img src="diagram.jpg" alt="Process Diagram">
    <figcaption>Figure 1: Our development process</figcaption>
</figure>
```

### Tables
```html
<table>
    <caption>Employee Information</caption>
    <thead>
        <tr>
            <th>Name</th>
            <th>Position</th>
            <th>Department</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td>John Doe</td>
            <td>Developer</td>
            <td>Engineering</td>
        </tr>
        <tr>
            <td>Jane Smith</td>
            <td>Designer</td>
            <td>Design</td>
        </tr>
    </tbody>
    <tfoot>
        <tr>
            <td colspan="3">Total Employees: 2</td>
        </tr>
    </tfoot>
</table>
```

### Forms
```html
<form action="/submit" method="post">
    <!-- Text input -->
    <label for="name">Name:</label>
    <input type="text" id="name" name="name" required>

    <!-- Email input -->
    <label for="email">Email:</label>
    <input type="email" id="email" name="email" required>

    <!-- Password input -->
    <label for="password">Password:</label>
    <input type="password" id="password" name="password" required>

    <!-- Number input -->
    <label for="age">Age:</label>
    <input type="number" id="age" name="age" min="1" max="120">

    <!-- Date input -->
    <label for="birthdate">Birth Date:</label>
    <input type="date" id="birthdate" name="birthdate">

    <!-- Textarea -->
    <label for="message">Message:</label>
    <textarea id="message" name="message" rows="4" cols="50"></textarea>

    <!-- Select dropdown -->
    <label for="country">Country:</label>
    <select id="country" name="country">
        <option value="">Select Country</option>
        <option value="us">United States</option>
        <option value="ca">Canada</option>
        <option value="uk">United Kingdom</option>
    </select>

    <!-- Radio buttons -->
    <fieldset>
        <legend>Gender:</legend>
        <input type="radio" id="male" name="gender" value="male">
        <label for="male">Male</label>
        <input type="radio" id="female" name="gender" value="female">
        <label for="female">Female</label>
    </fieldset>

    <!-- Checkboxes -->
    <fieldset>
        <legend>Interests:</legend>
        <input type="checkbox" id="coding" name="interests" value="coding">
        <label for="coding">Coding</label>
        <input type="checkbox" id="design" name="interests" value="design">
        <label for="design">Design</label>
    </fieldset>

    <!-- File upload -->
    <label for="file">Upload File:</label>
    <input type="file" id="file" name="file" accept=".jpg,.png,.pdf">

    <!-- Submit button -->
    <button type="submit">Submit</button>

    <!-- Reset button -->
    <button type="reset">Reset</button>
</form>
```

## Introduction to CSS

CSS (Cascading Style Sheets) controls the presentation and layout of HTML documents.

### What is CSS?
- **Style Language**: Defines how HTML elements are displayed
- **Separation of Concerns**: Separates content (HTML) from presentation (CSS)
- **Cascading**: Styles can inherit and override each other
- **Responsive**: Adapts to different screen sizes

### Ways to Add CSS

#### Inline Styles
```html
<p style="color: red; font-size: 20px;">Red text</p>
```

#### Internal Stylesheet
```html
<head>
    <style>
        p {
            color: blue;
            font-size: 16px;
        }
    </style>
</head>
```

#### External Stylesheet
```html
<head>
    <link rel="stylesheet" href="styles.css">
</head>
```

## CSS Selectors

Selectors target HTML elements to apply styles.

### Basic Selectors
```css
/* Element selector */
p {
    color: blue;
}

/* Class selector */
.highlight {
    background-color: yellow;
}

/* ID selector */
#header {
    font-size: 24px;
}

/* Universal selector */
* {
    margin: 0;
    padding: 0;
}
```

### Combinators
```css
/* Descendant selector */
div p {
    color: green;
}

/* Child selector */
div > p {
    font-weight: bold;
}

/* Adjacent sibling selector */
h1 + p {
    margin-top: 0;
}

/* General sibling selector */
h1 ~ p {
    color: gray;
}
```

### Attribute Selectors
```css
/* Elements with specific attribute */
input[type="text"] {
    border: 1px solid #ccc;
}

/* Attribute contains word */
[class*="btn"] {
    padding: 10px;
}

/* Attribute starts with */
[href^="https"] {
    color: green;
}

/* Attribute ends with */
[src$=".jpg"] {
    border: 2px solid black;
}
```

### Pseudo-Classes
```css
/* Link states */
a:link { color: blue; }
a:visited { color: purple; }
a:hover { color: red; }
a:active { color: orange; }

/* Form states */
input:focus { border-color: blue; }
input:valid { border-color: green; }
input:invalid { border-color: red; }

/* Position-based */
li:first-child { font-weight: bold; }
li:last-child { margin-bottom: 0; }
li:nth-child(even) { background-color: #f0f0f0; }

/* Content-based */
p:empty { display: none; }
```

### Pseudo-Elements
```css
/* First line of paragraph */
p::first-line {
    font-weight: bold;
}

/* First letter of paragraph */
p::first-letter {
    font-size: 150%;
}

/* Before content */
.quote::before {
    content: '"';
    font-size: 24px;
}

/* After content */
.quote::after {
    content: '"';
    font-size: 24px;
}

/* Selection */
::selection {
    background-color: yellow;
}
```

## CSS Properties

### Text Properties
```css
p {
    /* Font */
    font-family: Arial, sans-serif;
    font-size: 16px;
    font-weight: normal; /* normal, bold, 100-900 */
    font-style: normal; /* normal, italic, oblique */

    /* Text */
    color: #333;
    text-align: left; /* left, right, center, justify */
    text-decoration: none; /* none, underline, overline, line-through */
    text-transform: none; /* none, uppercase, lowercase, capitalize */
    letter-spacing: 0;
    word-spacing: 0;
    line-height: 1.5;

    /* Indentation */
    text-indent: 20px;
}
```

### Box Model
```css
div {
    /* Content dimensions */
    width: 200px;
    height: 100px;

    /* Padding (space inside border) */
    padding-top: 10px;
    padding-right: 20px;
    padding-bottom: 10px;
    padding-left: 20px;
    /* Shorthand: padding: 10px 20px; */

    /* Border */
    border-width: 1px;
    border-style: solid; /* none, solid, dashed, dotted, double */
    border-color: #000;
    /* Shorthand: border: 1px solid #000; */

    /* Margin (space outside border) */
    margin-top: 10px;
    margin-right: auto;
    margin-bottom: 10px;
    margin-left: auto;
    /* Shorthand: margin: 10px auto; */
}
```

### Background Properties
```css
.element {
    /* Background color */
    background-color: #f0f0f0;

    /* Background image */
    background-image: url('image.jpg');

    /* Background repeat */
    background-repeat: no-repeat; /* repeat, repeat-x, repeat-y, no-repeat */

    /* Background position */
    background-position: center top; /* left, center, right + top, center, bottom */

    /* Background size */
    background-size: cover; /* auto, cover, contain, or specific size */

    /* Background attachment */
    background-attachment: fixed; /* scroll, fixed, local */

    /* Shorthand */
    background: #f0f0f0 url('image.jpg') no-repeat center top / cover;
}
```

### Layout Properties

#### Display
```css
/* Block elements */
div {
    display: block; /* Takes full width, stacks vertically */
}

/* Inline elements */
span {
    display: inline; /* Flows with text, only takes needed width */
}

/* Inline-block */
.button {
    display: inline-block; /* Like inline but can have width/height */
}

/* None (hides element) */
.hidden {
    display: none;
}

/* Flexbox container */
.container {
    display: flex;
}

/* Grid container */
.grid {
    display: grid;
}
```

#### Position
```css
/* Static (default) */
.static {
    position: static;
}

/* Relative */
.relative {
    position: relative;
    top: 10px;
    left: 20px;
}

/* Absolute */
.absolute {
    position: absolute;
    top: 0;
    right: 0;
}

/* Fixed */
.fixed {
    position: fixed;
    bottom: 20px;
    right: 20px;
}

/* Sticky */
.sticky {
    position: sticky;
    top: 0;
}
```

#### Float and Clear
```css
/* Float */
.left-float {
    float: left;
    width: 50%;
}

.right-float {
    float: right;
    width: 50%;
}

/* Clear floats */
.clearfix::after {
    content: "";
    display: table;
    clear: both;
}
```

## CSS Flexbox

Flexbox is a modern layout system for creating flexible, responsive layouts.

### Flex Container
```css
.container {
    display: flex;

    /* Direction */
    flex-direction: row; /* row, row-reverse, column, column-reverse */

    /* Wrapping */
    flex-wrap: nowrap; /* nowrap, wrap, wrap-reverse */

    /* Shorthand */
    flex-flow: row wrap;

    /* Alignment */
    justify-content: flex-start; /* flex-start, flex-end, center, space-between, space-around, space-evenly */
    align-items: stretch; /* stretch, flex-start, flex-end, center, baseline */
    align-content: stretch; /* stretch, flex-start, flex-end, center, space-between, space-around */
}
```

### Flex Items
```css
.item {
    /* Size */
    flex-basis: auto;
    flex-grow: 0; /* How much it grows */
    flex-shrink: 1; /* How much it shrinks */

    /* Shorthand */
    flex: 1 1 auto; /* flex-grow flex-shrink flex-basis */

    /* Alignment */
    align-self: auto; /* auto, flex-start, flex-end, center, baseline, stretch */

    /* Order */
    order: 0; /* Lower numbers come first */
}
```

## CSS Grid

CSS Grid is a powerful 2D layout system.

### Grid Container
```css
.container {
    display: grid;

    /* Define columns */
    grid-template-columns: 1fr 2fr 1fr; /* Three columns */
    grid-template-columns: repeat(3, 1fr); /* Three equal columns */
    grid-template-columns: 100px auto 200px; /* Fixed and flexible */

    /* Define rows */
    grid-template-rows: 100px 200px;

    /* Gap between items */
    gap: 10px; /* Shorthand for row-gap and column-gap */
    row-gap: 10px;
    column-gap: 20px;

    /* Alignment */
    justify-items: start; /* start, end, center, stretch */
    align-items: start; /* start, end, center, stretch */
    justify-content: start; /* start, end, center, space-between, space-around, space-evenly */
    align-content: start; /* start, end, center, space-between, space-around, space-evenly */
}
```

### Grid Items
```css
.item {
    /* Position items */
    grid-column-start: 1;
    grid-column-end: 3;
    grid-row-start: 1;
    grid-row-end: 2;

    /* Shorthand */
    grid-column: 1 / 3; /* start / end */
    grid-row: 1 / 2;

    /* Span shorthand */
    grid-column: span 2; /* Span 2 columns */

    /* Named areas */
    grid-area: header; /* Reference to grid-template-areas */
}
```

## Responsive Design

### Media Queries
```css
/* Mobile first */
.container {
    width: 100%;
}

/* Tablet */
@media (min-width: 768px) {
    .container {
        width: 750px;
    }
}

/* Desktop */
@media (min-width: 992px) {
    .container {
        width: 970px;
    }
}

/* Large desktop */
@media (min-width: 1200px) {
    .container {
        width: 1170px;
    }
}
```

### Responsive Images
```css
img {
    max-width: 100%;
    height: auto;
}
```

### Flexible Layouts
```css
/* Fluid typography */
body {
    font-size: 16px;
}

@media (min-width: 768px) {
    body {
        font-size: 18px;
    }
}

/* Flexible containers */
.container {
    width: 100%;
    max-width: 1200px;
    margin: 0 auto;
    padding: 0 20px;
}
```

## CSS Best Practices

### Organization
```css
/* Reset/default styles */
* {
    box-sizing: border-box;
    margin: 0;
    padding: 0;
}

/* Base styles */
body {
    font-family: Arial, sans-serif;
    line-height: 1.6;
    color: #333;
}

/* Layout styles */
.header {
    background-color: #333;
    color: white;
    padding: 1rem;
}

.main-content {
    padding: 2rem;
}

/* Component styles */
.button {
    display: inline-block;
    padding: 0.5rem 1rem;
    background-color: #007bff;
    color: white;
    text-decoration: none;
    border-radius: 4px;
}

.button:hover {
    background-color: #0056b3;
}

/* Utility classes */
.text-center {
    text-align: center;
}

.hidden {
    display: none;
}
```

### Performance
- Minimize CSS file size
- Use CSS sprites for small images
- Avoid deep selector chains
- Use efficient selectors
- Minimize repaints and reflows

### Maintainability
- Use consistent naming conventions
- Comment your CSS
- Organize styles logically
- Use CSS variables for reusable values
- Avoid !important when possible

## Next Steps

You've learned the fundamentals of HTML and CSS! Here's what to explore next:

1. **JavaScript**: Add interactivity to your websites
2. **CSS Frameworks**: Bootstrap, Tailwind CSS, or Bulma
3. **Preprocessors**: Sass/SCSS for more powerful CSS
4. **Build Tools**: Webpack, Parcel, or Vite
5. **Version Control**: Git for tracking changes
6. **Deployment**: Host your websites online

Practice by building small projects and gradually increase complexity. The web development community is vast and welcoming!