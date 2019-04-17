angular.module('GruiApp').service('questionService', [
  "MainService",
  function questionService(MainService) {
    return {
      saveQuestion: function(data) {
        return MainService.post('/add-question', data);
      },

      editQuestion: function(data) {
        return MainService.put('/question/' + data.uid, data);
      },

      editScore: function(data) {
        return MainService.post('/question/editScore', data);
      },

      getAllQuestions: function(hideLoader) {
        return MainService.post('/get-all-questions', {}, hideLoader)
          .then(function(data) {
            if (!data || !data.data) {
              return [];
            }
            data = data.data
            var questions = data.questions || [];
            var answers = data.answers || [];

            questions.forEach(mainVm.fixQuestionUnescape);

            var questionUids = questions.reduce(function(acc, q) {
              acc[q.uid] = q;
              q.answerCount = 0;
              q.answerTotalScore = 0;
              q.skipCount = 0;
              return acc
            }, {})

            answers.forEach(function(answer) {
              var question = questionUids[answer.questionUid];
              if (!question) {
                console.error('Uknown question for answer ', answer);
                return
              }
              question.answerCount = answer.totalCount;
              question.answerTotalScore = answer.totalScore;
              question.skipCount = answer.skippedCount || 0;
              question.difficulty = answer.correctCount / answer.totalCount;
            })

            return questions;
          })
      },

      getQuestion: function(questionId) {
        return MainService.get('/question/' + questionId)
          .then(function(data) {
            return mainVm.fixQuestionUnescape(data.data.question[0]);
          })
      },

      updateAllScores: function() {
        var query = `{
        	quiz(func: has(quiz.candidate)) {
        		candidate: quiz.candidate {
        			uid
        			name
        			score
        			complete
        			deleted
        			quiz_start
        			candidate.question {
                uid
                question {
                  uid
                  positive
                  negative
                  question.correct {
                    uid
                  }
                }
        				candidate.answer
                candidate.score
        			}
        		}
        	}
        }`;

        return MainService.proxy(query).then(function(resp) {
          const allCandidates = resp.data.quiz
              .map(quiz => quiz.candidate)
              .reduce((acc, x) => acc.concat(x), [])

          var mutation = [];
          var legacy = [];

          allCandidates.forEach(cand => {
            if (!cand.complete || cand.deleted) {
              return
            }
            var totalScore = 0;
            var isLegacy = false;
            cand["candidate.question"].forEach(q => {
              // console.log('q=',q)
              var question = q.question[0]
              var answers = (q['candidate.answer'] || '').split(',')
              var newScore = 0
              if (answers.length && answers[0] && answers[0] !== 'skip') {
                if (answers[0].length > 9) {
                  console.log('legacy answer: ', q)
                  isLegacy = true;
                  return;
                }
                answers.forEach(ans => {
                  var isCorrect = question['question.correct'].map(q => q.uid).indexOf(ans) >= 0
                  newScore += isCorrect ? question.positive : -question.negative
                })
              }
              totalScore += newScore;
              if (q['candidate.score'] !== newScore) {
                mutation.push(`<${q.uid}> <candidate.score> "${newScore}" .`)
              }
            })
            if (cand.score !== totalScore && !isLegacy) {
              mutation.push(`<${cand.uid}> <score> "${totalScore}" .`)
            }
          });

          console.log(mutation.join('\n'))
          return MainService.mutateProxy(`{
              set {
                ${mutation.join('\n')}
              }
            }`)
        });
      }
    }
  }
]);

angular.module('GruiApp').service('allQuestions', [
  'questionService',
  '$rootScope',
  function(questionService, $rootScope) {
    var allQuestions = [];

    function fetchQuestions() {
      questionService.getAllQuestions(true).then(
        function(questions) {
          setTimeout(function() {
            $rootScope.$apply(function() {
              allQuestions = questions;
            });
          }, 1);
        },
        function(err) {
          console.error(err);
        });
    }
    fetchQuestions();

    setInterval(fetchQuestions, 60000);

    return {
      get: function() {
        return allQuestions;
      },
      refresh: fetchQuestions,
    }
  }
]);
